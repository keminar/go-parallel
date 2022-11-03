package parallel

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// 测试替换工人
func TestReplace(*testing.T) {
	p := NewParallelPool(10, 2)
	defer p.Close()
	for i := 0; i < 10; i++ {
		p.AddTask(i)
	}
	for j := 0; j < 2; j++ {
		p.AddWorker(j)
	}
	p.Run(func(worker int, task int) bool {
		fmt.Println("TestReplace ", task, "worker=", worker, time.Now().Unix())
		if worker == 0 {
			time.Sleep(time.Second * 5)
		} else {
			time.Sleep(time.Second * 1)
		}
		if worker == 1 {
			p.ReplaceWorker(1, 4)
			// 再替换一次失败的
			p.ReplaceWorker(1, 3)
		}
		return false
	})
}

// 测试重复添加工人
func TestRepeatAdd(*testing.T) {
	total := 10
	workerNum := 2
	p := NewParallelPool(total, workerNum)
	defer p.Close()
	for i := 0; i < total; i++ {
		p.AddTask(i)
	}
	for i := 0; i < workerNum; i++ {
		p.AddWorker(i)
	}
	// addworker repeat
	p.AddWorker(0)
	p.Run(func(worker int, task int) bool {
		fmt.Println("TestRepeatAdd ", task, "worker=", worker, time.Now().Unix())
		time.Sleep(time.Second * 1)
		p.ReplaceWorker(0, 3)
		return false
	})
}

// 测试不同返回值
func TestReturn(*testing.T) {
	total := 10
	workerNum := 2
	p := NewParallelPool(total, workerNum)
	defer p.Close()
	for i := 0; i < total; i++ {
		p.AddTask(i)
	}
	for i := 0; i < workerNum; i++ {
		p.AddWorker(i)
	}
	p.Run(func(worker int, task int) bool {
		time.Sleep(time.Second * 1)
		fmt.Println("TestReturn ", task, "worker=", worker)
		if worker%2 == 0 {
			return true
		}
		return false
	})
}

// 测试崩溃
func TestPanic(t *testing.T) {
	total := 10
	workerNum := 2
	p := NewParallelPool(total, workerNum)
	defer p.Close()
	for i := 0; i < total; i++ {
		p.AddTask(i)
	}
	for i := 0; i < workerNum; i++ {
		p.AddWorker(i)
	}
	p.Run(func(worker int, task int) bool {
		time.Sleep(time.Second * 1)
		var x interface{}
		x = nil
		fmt.Println("TestPanic ", task, "worker=", worker, "111")
		var de struct{}
		json.Unmarshal(x.([]byte), &de)
		fmt.Println("TestPanic ", task, "worker=", worker, "222")
		return false
	})
}
