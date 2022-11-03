# go-parallel
比waitGroup更智能的go并行处理库

# 使用
```
  p := NewParallelPool(10, 2)
	for i := 0; i < 10; i++ {
		p.AddTask(i)
	}
	for j := 0; j < 2; j++ {
		p.AddWorker(j)
	}
	p.Run(func(worker int, task int) bool {
		 fmt.Println("TestReplace ", task, "worker=", worker, time.Now().Unix())
	   time.Sleep(time.Second * 1) 
		 return false
	})
```
