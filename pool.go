package parallel

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

type parallelPool struct {
	wg          sync.WaitGroup
	taskNum     int
	taskQueue   chan int
	workers     *workerMap
	workerQueue chan int
	sleep       time.Duration
	logPrint    func(v ...interface{})
}

func NewParallelPool(taskNum int, workerNum int) *parallelPool {
	p := &parallelPool{
		wg:          sync.WaitGroup{},
		taskNum:     taskNum,
		taskQueue:   make(chan int, taskNum),
		workers:     newWorkerMap(),
		workerQueue: make(chan int, workerNum),
		sleep:       0,
		logPrint:    log.Println,
	}
	p.wg.Add(taskNum)
	return p
}

// 并发停顿
func (p *parallelPool) SetSleep() {
	p.sleep = time.Duration(10) * time.Millisecond
}

// 传输任务数
func (p *parallelPool) AddTask(task int) {
	p.taskQueue <- task
}

// 添加工人, 相同工人重复添加无效
func (p *parallelPool) AddWorker(worker int) {
	p.workers.Lock()
	defer p.workers.Unlock()

	defer func() {
		// 捕获异常
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			p.logPrint(fmt.Sprintf("panic2: %v\n%s", err, buf))
		}
	}()

	if _, ok := p.workers.data[worker]; !ok {
		p.workers.data[worker] = struct{}{}
		// 一个工人只有一条生命, 不可对同一个工人多次加队列增加权重，否则会导致队列长度不够写入卡住
		// 队列值要和map没有一一对应的另一个坏处是还会导致并发工作比工人数多，且在替换工人时并发数量又降低
		p.workerQueue <- worker
	}
}

// 替换新工人, 返回false为已有工人，true为替换工人成功
// 最新代码已经不关心此函数返回值。Run不拿这个返回值做判断了
func (p *parallelPool) ReplaceWorker(last, newer int) bool {
	if last == newer { //同一个人，不算替换
		return false
	}
	p.workers.Lock()
	defer p.workers.Unlock()
	if _, ok := p.workers.data[last]; !ok {
		return false
	}
	if _, ok := p.workers.data[newer]; !ok {
		defer func() {
			// 捕获异常
			if err := recover(); err != nil {
				p.logPrint(fmt.Sprintf("panic: %v", err))
			} else {
				delete(p.workers.data, last)
			}
		}()
		p.workers.data[newer] = struct{}{}
		p.workerQueue <- newer
		return true
	}
	return false
}

// 设置工人为可用
func (p *parallelPool) setWorker(worker int) {
	p.workers.Lock()
	defer p.workers.Unlock()

	defer func() {
		// 捕获异常
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			p.logPrint(fmt.Sprintf("panic2: %v\n%s", err, buf))
		}
	}()

	// 只要当前工人还在，说明没有进行过替换掉此工人，则需要加回
	// 如果工人不在说明通过ReplaceWorker加了新的工人，这边不需要再操作
	if _, ok := p.workers.data[worker]; ok {
		p.workerQueue <- worker
	}
}

// 设置日志输出
func (p *parallelPool) SetLogger(name func(v ...interface{})) {
	p.logPrint = name
}

// 传输，并执行回调
// 最新代码已经不关心fn的返回值
func (p *parallelPool) Run(fn func(worker int, task int) bool) {
	for i := 0; i < p.taskNum; i++ {
		go func() {
			// worker被吃空后就会等待之前的处理完释放
			w := <-p.workerQueue
			t := <-p.taskQueue

			defer func() {
				// 捕获异常 防止waitGroup阻塞
				if err := recover(); err != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					p.logPrint(fmt.Sprintf("panic: %v\n%s", err, buf))
				}

				// 将下面的代码放在defer是为了防止fn出现panic时工人没有被放回，导致没工人卡住
				// 为防止worker被错误的减成空，这里保底一个工人
				if p.workers.length() == 0 {
					p.AddWorker(w)
				} else {
					p.setWorker(w)
				}
				// 这个休息时间不应该放go外面，只是让addWorker/setWorker延迟一点
				if p.sleep > 0 {
					time.Sleep(p.sleep)
				}
				// 放在最末尾，保证在setWorker前chan不被close
				p.wg.Done()
			}()

			// 当节点有变更时在fn函数可能会更新worker
			// 如果fn出现panic，工人也要放回
			// 返回值以前用来判断是否有替换工作，现废弃不再使用，安全起见先不删除调用定义
			_ = fn(w, t)
		}()
	}
	p.wg.Wait()
}

// 释放资源
func (p *parallelPool) Close() {
	close(p.taskQueue)
	close(p.workerQueue)
}
