package seed

import (
	"codegym/backend/internal/model"
)

// toGoCSeed converts a Go concurrency seed under the "并发" subcategory.
func toGoCSeed(g goBasicSeed) seedExercise {
	ex := toGoSeed(g)
	ex.subcategory = "并发"
	return ex
}

// GoConcurrency returns the 50-question Go concurrency bank (并发).
func GoConcurrency() ([]model.Exercise, error) {
	pt := func(id, text, category, weight, hint string) seedPoint {
		return seedPoint{id: id, text: text, category: category, minWeight: weight,
			weightScore: 1, hint: hint, fillType: "text"}
	}
	fill := func(id, title, desc, category, weight, code, expected string, points []seedPoint, l4, l1, l2, l3, expl string) goBasicSeed {
		return goBasicSeed{id: id, title: title, desc: desc, category: category, weight: weight,
			kind: "fill", code: code, expected: expected, points: points,
			hintL1: l1, hintL2: l2, hintL3: l3, hintL4: l4, explanation: expl}
	}
	bugfix := func(id, title, desc, category, weight, code, buggy, expected string, tcs []model.TestCase, p seedPoint, l4, l1, l2, l3, expl string) goBasicSeed {
		return goBasicSeed{id: id, title: title, desc: desc, category: category, weight: weight,
			kind: "bugfix", code: code, buggy: buggy, expected: expected, testCases: tcs,
			points: []seedPoint{p}, hintL1: l1, hintL2: l2, hintL3: l3, hintL4: l4, explanation: expl}
	}

	seeds := []goBasicSeed{
		// ---- 入门：goroutine 与基础等待 ----
		fill("go-cc-001", "启动 goroutine", "用 go 关键字启动协程并等待", "goroutine", "beginner",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\twg.Add(1)\n\tgo func() {\n\t\tdefer wg.Done()\n\t\tfmt.Println(\"hi\")\n\t}()\n\twg.Wait()\n}",
			"hi",
			[]seedPoint{pt("go", "go func", "keyword", "beginner", "启动协程的关键字"), pt("wait", "wg.Wait()", "waitgroup", "easy", "等待协程完成")},
			"`go func` / `wg.Wait()`", "用 go 启动协程", "`go` 启动协程，`WaitGroup` 等它完成", "`go func(){ ... }()`",
			"`go` 关键字并发执行函数；主协程用 `WaitGroup.Wait()` 等待。"),

		fill("go-cc-002", "WaitGroup 等待", "启动多个协程并等待全部完成", "waitgroup", "beginner",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tdone := make(chan int, 5)\n\tfor i := 0; i < 5; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tdone <- 1\n\t\t}()\n\t}\n\twg.Wait()\n\tclose(done)\n\ts := 0\n\tfor range done {\n\t\ts++\n\t}\n\tfmt.Println(s)\n}",
			"5",
			[]seedPoint{pt("add", "wg.Add(1)", "waitgroup", "beginner", "每次启动前计数 +1"), pt("done", "done <- 1", "channel", "beginner", "协程完成后发信号")},
			"`wg.Add(1)` / `done <- 1`", "WaitGroup + channel", "`Add/Done/Wait` 等待全部协程，channel 汇总结果", "`wg.Wait()`",
			"WaitGroup 等待 5 个协程全部 `Done`，再用 channel 收齐，故 s=5。"),

		fill("go-cc-003", "goroutine 参数捕获", "协程通过参数传入循环变量", "goroutine", "easy",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tch := make(chan int, 3)\n\tfor i := 1; i <= 3; i++ {\n\t\twg.Add(1)\n\t\tgo func(n int) {\n\t\t\tdefer wg.Done()\n\t\t\tch <- n\n\t\t}(i)\n\t}\n\twg.Wait()\n\tclose(ch)\n\ts := 0\n\tfor v := range ch {\n\t\ts += v\n\t}\n\tfmt.Println(s)\n}",
			"6",
			[]seedPoint{pt("param", "go func(n int)", "closure", "easy", "把循环变量作为参数传入"), pt("call", "(i)", "closure", "medium", "传入当前 i 的值")},
			"`func(n int)` / `(i)`", "按值传参捕获", "把 i 作为参数传入，避免共享循环变量", "`go func(n int){...}(i)`",
			"把循环变量作为参数传入协程，每个协程拥有独立副本，结果 1+2+3=6。"),

		fill("go-cc-004", "无缓冲 channel 通信", "协程用 channel 把值传给主协程", "channel", "beginner",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int)\n\tgo func() { ch <- 7 }()\n\tv := <-ch\n\tfmt.Println(v)\n}",
			"7",
			[]seedPoint{pt("send", "ch <- 7", "channel", "beginner", "发送值"), pt("recv", "<-ch", "channel", "beginner", "接收值")},
			"`ch <- 7` / `<-ch`", "无缓冲 channel", "发送后接收方就绪才完成（同步）", "`ch <- 7` → `v := <-ch`",
			"`ch <- v` 发送、`v := <-ch` 接收；无缓冲 channel 是同步的。"),

		fill("go-cc-005", "缓冲 channel", "用带缓冲的 channel 避免阻塞", "channel", "beginner",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 3)\n\tch <- 1\n\tch <- 2\n\tch <- 3\n\tfmt.Println(len(ch))\n}",
			"3",
			[]seedPoint{pt("make", "make(chan int, 3)", "channel", "beginner", "容量为 3 的缓冲")},
			"`make(chan int, 3)`", "缓冲 channel", "容量未满时发送不阻塞", "`len(ch)` 是已存数量",
			"`make(chan T, n)` 创建容量 n 的缓冲；未满前发送不阻塞。"),

		fill("go-cc-006", "遍历 channel 求和", "发送后 close 再 range 求和", "channel", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 3)\n\tch <- 1\n\tch <- 2\n\tch <- 3\n\tclose(ch)\n\ts := 0\n\tfor v := range ch {\n\t\ts += v\n\t}\n\tfmt.Println(s)\n}",
			"6",
			[]seedPoint{pt("close", "close(ch)", "channel", "easy", "关闭 channel 结束 range")},
			"`close(ch)`", "关闭 channel", "`range ch` 在关闭后结束", "`for v := range ch`",
			"发送完 `close(ch)` 后 `range` 才结束；不关闭则无限等待。"),

		fill("go-cc-007", "只发送 channel", "用只写 channel 参数", "channel_direction", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc send(ch chan<- int) {\n\tch <- 42\n}\n\nfunc main() {\n\tch := make(chan int)\n\tgo send(ch)\n\tfmt.Println(<-ch)\n}",
			"42",
			[]seedPoint{pt("dir", "chan<- int", "channel_direction", "medium", "只写的 channel 类型")},
			"`chan<- int`", "单向 channel", "`chan<-` 只写、`<-chan` 只读", "`func send(ch chan<- int)`",
			"`chan<- T` 声明只写 channel，编译器在函数内禁止接收。"),

		fill("go-cc-008", "只接收 channel", "用只读 channel 参数", "channel_direction", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc recv(ch <-chan int) int {\n\treturn <-ch\n}\n\nfunc main() {\n\tch := make(chan int, 1)\n\tch <- 7\n\tfmt.Println(recv(ch))\n}",
			"7",
			[]seedPoint{pt("dir", "<-chan int", "channel_direction", "medium", "只读的 channel 类型")},
			"`<-chan int`", "单向只读", "`<-chan T` 只读，防止误发送", "`func recv(ch <-chan int)`",
			"`<-chan T` 只读 channel；把 channel 按方向限制传给函数更安全。"),

		fill("go-cc-009", "接收 ok 判断", "从已关闭 channel 接收并判断状态", "channel", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 1)\n\tch <- 5\n\tclose(ch)\n\tv, ok := <-ch\n\tfmt.Println(v, ok)\n}",
			"5 true",
			[]seedPoint{pt("ok", "ok", "channel", "easy", "是否成功收到")},
			"`v, ok := <-ch`", "ok 判断", "ok=false 表示 channel 已关闭且无值", "`fmt.Println(v, ok)`",
			"接收返回 `(值, ok)`，ok=false 表示已关闭无数据。"),

		fill("go-cc-010", "select 就绪分支", "用 select 从有值的 channel 接收", "select", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\ta := make(chan int, 1)\n\tb := make(chan int, 1)\n\tb <- 2\n\tselect {\n\tcase x := <-a:\n\t\tfmt.Println(\"a\", x)\n\tcase x := <-b:\n\t\tfmt.Println(\"b\", x)\n\t}\n}",
			"b 2",
			[]seedPoint{pt("select", "select", "keyword", "easy", "多路复用关键字"), pt("case", "case x := <-b:", "select", "medium", "接收 b 的分支")},
			"`select` / `case x := <-b:`", "select 接收", "b 有值则走该分支", "`select { case x := <-b: ... }`",
			"`select` 多个 case 中随选一个就绪的（都有值时随机）。"),

		fill("go-cc-011", "select default", "无就绪分支时走 default", "select", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int)\n\tselect {\n\tcase v := <-ch:\n\t\tfmt.Println(v)\n\tdefault:\n\t\tfmt.Println(\"default\")\n\t}\n}",
			"default",
			[]seedPoint{pt("default", "default:", "select", "easy", "无就绪分支的默认项")},
			"`default:`", "select default", "所有 case 未就绪时走 default", "`select { default: }`",
			"`select` 的 `default` 在其它 case 都未就绪时立即执行，用于非阻塞探测。"),

		fill("go-cc-012", "select 超时", "用 time.After 实现超时", "select", "medium",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"time\"\n)\n\nfunc main() {\n\tch := make(chan int)\n\tselect {\n\tcase v := <-ch:\n\t\tfmt.Println(v)\n\tcase <-time.After(10 * time.Millisecond):\n\t\tfmt.Println(\"timeout\")\n\t}\n}",
			"timeout",
			[]seedPoint{pt("after", "time.After(10 * time.Millisecond)", "timeout", "medium", "定时通道")},
			"`time.After`", "select 超时", "time.After 到期后可读，实现超时", "`case <-time.After(10ms):`",
			"`time.After(d)` 到点后向其通道发送，配合 select 做超时。"),

		// ---- sync 原语 ----
		fill("go-cc-013", "Mutex 互斥", "用 Mutex 保护共享计数", "mutex", "medium",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tvar mu sync.Mutex\n\tn := 0\n\tfor i := 0; i < 1000; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tmu.Lock()\n\t\t\tn++\n\t\t\tmu.Unlock()\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(n)\n}",
			"1000",
			[]seedPoint{pt("lock", "mu.Lock()", "mutex", "medium", "加锁"), pt("unlock", "mu.Unlock()", "mutex", "medium", "解锁")},
			"`mu.Lock()` / `mu.Unlock()`", "Mutex 保护临界区", "加锁后再读写共享变量", "`mu.Lock(); n++; mu.Unlock()`",
			"`Lock/Unlock` 保证同一时刻只有一个协程进入临界区。"),

		fill("go-cc-014", "RWMutex 读写锁", "多个读并发、写互斥", "rwmutex", "hard",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar mu sync.RWMutex\n\tvar wg sync.WaitGroup\n\treads := 0\n\tfor i := 0; i < 5; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tmu.RLock()\n\t\t\t_ = reads\n\t\t\tmu.RUnlock()\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(reads)\n}",
			"0",
			[]seedPoint{pt("rlock", "mu.RLock()", "rwmutex", "hard", "读锁")},
			"`mu.RLock()`", "读写锁", "RLock 共享读，写 Lock 独占", "`mu.RLock() ... mu.RUnlock()`",
			"`RWMutex` 允许多个读并发；写者需要独占 `Lock`。"),

		fill("go-cc-015", "sync.Once 只执行一次", "多个协程只初始化一次", "once", "medium",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar once sync.Once\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 5; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tonce.Do(func() { fmt.Println(\"init\") })\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(\"done\")\n}",
			"init\ndone",
			[]seedPoint{pt("once", "once.Do", "once", "medium", "只执行一次的方法")},
			"`once.Do(...)`", "sync.Once", "无论多少协程调用都只执行一次", "`once := sync.Once{}`",
			"`Once.Do(f)` 保证 f 只执行一次，即使多个协程并发调用。"),

		fill("go-cc-016", "atomic 原子加", "用 atomic 安全计数", "atomic", "medium",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n\t\"sync/atomic\"\n)\n\nfunc main() {\n\tvar n int64\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 100; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tatomic.AddInt64(&n, 1)\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(n)\n}",
			"100",
			[]seedPoint{pt("add", "atomic.AddInt64(&n, 1)", "atomic", "medium", "原子自增")},
			"`atomic.AddInt64(&n, 1)`", "原子操作", "无需加锁也能安全累加", "`atomic.AddInt64(&n, 1)`",
			"`sync/atomic` 提供无锁原子运算，计数器并发安全。"),

		fill("go-cc-017", "CAS 原子比较交换", "用 CompareAndSwap 更新值", "atomic", "hard",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync/atomic\"\n)\n\nfunc main() {\n\tvar n int64 = 5\n\tok := atomic.CompareAndSwapInt64(&n, 5, 10)\n\tfmt.Println(n, ok)\n}",
			"10 true",
			[]seedPoint{pt("cas", "CompareAndSwapInt64(&n, 5, 10)", "atomic", "hard", "比较并交换")},
			"`CompareAndSwapInt64(&n, 5, 10)`", "CAS 原子", "值等于旧值才替换，返回是否成功", "`atomic.CompareAndSwapInt64(&n, 5, 10)`",
			"CAS：仅当当前值等于旧值时写入新值，失败可重试实现乐观更新。"),

		// ---- channel 模式 ----
		fill("go-cc-018", "channel 生成器", "用 channel 生成序列", "channel_pattern", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc nums() <-chan int {\n\tch := make(chan int)\n\tgo func() {\n\t\tfor i := 1; i <= 4; i++ {\n\t\t\tch <- i\n\t\t}\n\t\tclose(ch)\n\t}()\n\treturn ch\n}\n\nfunc main() {\n\ts := 0\n\tfor v := range nums() {\n\t\ts += v\n\t}\n\tfmt.Println(s)\n}",
			"10",
			[]seedPoint{pt("ret", "<-chan int", "channel_pattern", "medium", "返回只读 channel")},
			"`nums() <-chan int`", "生成器模式", "函数用协程生产、channel 传出", "`for v := range nums()`",
			"函数返回只读 channel，后台协程往里发送数据（生产者）。"),

		fill("go-cc-019", "Worker Pool 任务池", "多个 worker 处理任务", "channel_pattern", "hard",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc worker(id int, jobs <-chan int, res chan<- int, wg *sync.WaitGroup) {\n\tfor j := range jobs {\n\t\tres <- j * 2\n\t}\n\twg.Done()\n}\n\nfunc main() {\n\tjobs := make(chan int, 5)\n\tres := make(chan int, 5)\n\tvar wg sync.WaitGroup\n\tfor w := 0; w < 2; w++ {\n\t\twg.Add(1)\n\t\tgo worker(w, jobs, res, &wg)\n\t}\n\tfor j := 1; j <= 5; j++ {\n\t\tjobs <- j\n\t}\n\tclose(jobs)\n\twg.Wait()\n\tclose(res)\n\ts := 0\n\tfor v := range res {\n\t\ts += v\n\t}\n\tfmt.Println(s)\n}",
			"30",
			[]seedPoint{pt("range", "for j := range jobs", "channel_pattern", "hard", "从任务 channel 取任务"), pt("close", "close(jobs)", "channel_pattern", "hard", "投完任务后关闭")},
			"`for j := range jobs` / `close(jobs)`", "Worker Pool", "多个 worker 从 jobs channel 消费", "`for j := range jobs`",
			"固定 worker 并行消费任务 channel；作业投完 `close`。结果 2×(1..5)=30。"),

		fill("go-cc-020", "Fan-in 汇聚", "多个协程写入同一 channel", "channel_pattern", "medium",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tch := make(chan int, 2)\n\tvar wg sync.WaitGroup\n\tfor i := 1; i <= 2; i++ {\n\t\twg.Add(1)\n\t\tgo func(n int) {\n\t\t\tdefer wg.Done()\n\t\t\tch <- n\n\t\t}(i)\n\t}\n\twg.Wait()\n\tclose(ch)\n\ts := 0\n\tfor v := range ch {\n\t\ts += v\n\t}\n\tfmt.Println(s)\n}",
			"3",
			[]seedPoint{pt("close", "close(ch)", "channel_pattern", "medium", "全部写完后关闭")},
			"`close(ch)`", "Fan-in 汇聚", "多个生产者写一个 channel，主协程读", "`for v := range ch`",
			"Fan-in：多个协程向同一 channel 写入，主协程汇聚消费。"),

		fill("go-cc-021", "Pipeline 流水线", "两级管道传递并变换", "channel_pattern", "hard",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tin := make(chan int)\n\tout := make(chan int)\n\tgo func() {\n\t\tin <- 4\n\t\tclose(in)\n\t}()\n\tgo func() {\n\t\tfor v := range in {\n\t\t\tout <- v * 2\n\t\t}\n\t\tclose(out)\n\t}()\n\tfmt.Println(<-out)\n}",
			"8",
			[]seedPoint{pt("map", "v * 2", "channel_pattern", "hard", "在流水线中变换值"), pt("close", "close(out)", "channel_pattern", "hard", "输出完成关闭")},
			"`v * 2` / `close(out)`", "Pipeline 流水线", "阶段间用 channel 传值", "`out <- v * 2`",
			"Pipeline：上阶段写入、下阶段变换后再输出；4×2=8。"),

		fill("go-cc-022", "cap 与 len of channel", "查看 channel 容量与元素数", "channel", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 5)\n\tch <- 1\n\tch <- 2\n\tfmt.Println(len(ch), cap(ch))\n}",
			"2 5",
			[]seedPoint{pt("cap", "cap(ch)", "channel", "easy", "channel 容量")},
			"`cap(ch)`", "channel 容量", "len=已存，cap=容量", "`fmt.Println(len(ch), cap(ch))`",
			"`len` 已存数量、`cap` 总容量（仅缓冲 channel 有意义）。"),

		fill("go-cc-023", "close 广播", "关闭 channel 同时通知多个协程", "channel", "hard",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tdone := make(chan struct{})\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 3; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\t<-done\n\t\t\tfmt.Println(\"x\")\n\t\t}()\n\t}\n\tclose(done)\n\twg.Wait()\n}",
			"x\nx\nx",
			[]seedPoint{pt("close", "close(done)", "channel", "hard", "关闭即广播")},
			"`close(done)`", "close 广播", "所有等待接收的协程同时被唤醒", "`<-done` 解除阻塞",
			"关闭空 channel 会同时唤醒所有等待它的接收者，用作退出信号。"),

		fill("go-cc-024", "信号量限流", "用缓冲 channel 限制并发数", "pattern", "hard",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tsem := make(chan struct{}, 2)\n\tcnt := 0\n\tfor i := 0; i < 10; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tsem <- struct{}{}\n\t\t\tcnt++\n\t\t\t<-sem\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(cnt, cap(sem))\n}",
			"10 2",
			[]seedPoint{pt("sem", "sem <- struct{}{}", "pattern", "hard", "占一个信号量槽"), pt("rel", "<-sem", "pattern", "hard", "释放槽位")},
			"`sem <- struct{}{}` / `<-sem`", "信号量", "并发数=信号量容量", "`sem <- struct{}{}` 占用，`<-sem` 释放",
			"用容量受限的 channel 作信号量，`cnt++` 无锁不保证但此处并发数小。"),

		// ---- context ----
		fill("go-cc-025", "context 取消", "用 cancel 通知协程退出", "context", "hard",
			"package main\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tctx, cancel := context.WithCancel(context.Background())\n\tvar wg sync.WaitGroup\n\twg.Add(1)\n\tgo func() {\n\t\tdefer wg.Done()\n\t\tselect {\n\t\tcase <-ctx.Done():\n\t\t\tfmt.Println(\"canceled\")\n\t\t}\n\t}()\n\tcancel()\n\twg.Wait()\n}",
			"canceled",
			[]seedPoint{pt("done", "ctx.Done()", "context", "hard", "取消通知通道"), pt("cancel", "cancel()", "context", "hard", "触发取消")},
			"`ctx.Done()` / `cancel()`", "context 取消", "cancel 后 Done 关闭，select 解除阻塞", "`case <-ctx.Done():`",
			"`WithCancel` 派生子 context，调 `cancel()` 关闭 `Done()` 通知协程退出。"),

		fill("go-cc-026", "context 超时", "用 WithTimeout 自动取消", "context", "hard",
			"package main\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\t\"time\"\n)\n\nfunc main() {\n\tctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)\n\tdefer cancel()\n\tselect {\n\tcase <-ctx.Done():\n\t\tfmt.Println(\"timeout\")\n\t}\n}",
			"timeout",
			[]seedPoint{pt("to", "context.WithTimeout", "context", "hard", "带超时的 context")},
			"`context.WithTimeout`", "context 超时", "到点后 Done 自动关闭", "`case <-ctx.Done():`",
			"`WithTimeout` 到时时自动取消，`Done()` 会关闭并提示原因。"),

		fill("go-cc-027", "context 传值", "用 WithValue 携带请求级数据", "context", "hard",
			"package main\n\nimport (\n\t\"context\"\n\t\"fmt\"\n)\n\nfunc main() {\n\tctx := context.WithValue(context.Background(), \"user\", \"admin\")\n\tfmt.Println(ctx.Value(\"user\"))\n}",
			"admin",
			[]seedPoint{pt("value", "context.WithValue", "context", "hard", "携带键值")},
			"`context.WithValue` / `Value(\"user\")`", "context 传值", "键-值传给下游", "`ctx.Value(\"user\")`",
			"`WithValue` 在 context 链上携带请求级数据，用 `Value(key)` 读取。"),

		fill("go-cc-028", "done 通道退出", "用关闭的 done 通知退出", "channel", "medium",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tdone := make(chan struct{})\n\tvar wg sync.WaitGroup\n\twg.Add(1)\n\tgo func() {\n\t\tdefer wg.Done()\n\t\t<-done\n\t\tfmt.Println(\"exit\")\n\t}()\n\tclose(done)\n\twg.Wait()\n}",
			"exit",
			[]seedPoint{pt("recv", "<-done", "channel", "medium", "接收退出信号")},
			"`<-done`", "done 通道", "关闭 done 通知协程退出", "`close(done)`",
			"用单独的 done channel，关闭即发退出信号，避免协程泄漏。"),

		// ---- 生产者消费者 / 分发 ----
		fill("go-cc-029", "生产者消费者", "生产者投递、消费者处理", "pattern", "hard",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tch := make(chan int, 5)\n\tvar wg sync.WaitGroup\n\twg.Add(1)\n\tgo func() {\n\t\tdefer wg.Done()\n\t\tfor i := 1; i <= 5; i++ {\n\t\t\tch <- i\n\t\t}\n\t\tclose(ch)\n\t}()\n\ts := 0\n\tfor v := range ch {\n\t\ts += v\n\t}\n\twg.Wait()\n\tfmt.Println(s)\n}",
			"15",
			[]seedPoint{pt("close", "close(ch)", "pattern", "hard", "生产完关闭")},
			"`close(ch)`", "生产者消费者", "生产者写完关闭，消费者 range 读完", "`for v := range ch`",
			"生产者往缓冲 channel 投递并 close，消费者 range 消费，1..5 求和 15。"),

		// ---- bug 修复 ----
		bugfix("go-cc-030", "WaitGroup 死锁修复", "修复缺少 wg.Done() 导致的死锁", "bugfix", "medium",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 3; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tfmt.Println(\"x\")\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(\"done\")\n}",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 3; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tfmt.Println(\"x\")\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(\"done\")\n}",
			"x\nx\nx\ndone",
			[]model.TestCase{{Output: "done"}},
			pt("done", "fmt.Println(\"x\")", "waitgroup", "medium", "协程内输出"),
			"`defer wg.Done()`", "wg.Done 必须调用", "缺少 Done 计数永不归零，Wait 卡死", "`defer wg.Done()`",
			"每个协程结束都要 `Done()`；否则 `Wait` 永远阻塞（死锁）。"),

		bugfix("go-cc-031", "数据竞争修复", "给共享计数加锁修复数据竞争", "bugfix", "hard",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tvar mu sync.Mutex\n\tn := 0\n\tfor i := 0; i < 1000; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tmu.Lock()\n\t\t\tn++\n\t\t\tmu.Unlock()\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(n)\n}",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tn := 0\n\tfor i := 0; i < 1000; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tn++\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(n)\n}",
			"1000",
			[]model.TestCase{{Output: "1000"}},
			pt("lock", "n++", "mutex", "hard", "共享计数操作"),
			"`mu.Lock(); n++; mu.Unlock()`", "数据竞争", "多个协程无锁写同一变量会丢失更新", "`mu.Lock()` 保护 `n++`",
			"并发写共享变量需互斥；加锁保证 n 最终等于 1000。"),

		bugfix("go-cc-032", "缓冲不足死锁修复", "修复缓冲过小导致发送阻塞", "bugfix", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 3)\n\tch <- 1\n\tch <- 2\n\tch <- 3\n\tfmt.Println(len(ch))\n}",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int)\n\tch <- 1\n\tch <- 2\n\tch <- 3\n\tfmt.Println(len(ch))\n}",
			"3",
			[]model.TestCase{{Output: "3"}},
			pt("cap", "make(chan int)", "channel", "medium", "缓冲容量不足"),
			"`make(chan int, 3)`", "无缓冲发送阻塞", "无缓冲 channel 发送需接收方，否则阻塞", "`make(chan int, 3)`",
			"无缓冲 channel 发送会阻塞到有人接收；改带缓冲即可容纳 3 个值。"),

		bugfix("go-cc-033", "Once 重复初始化修复", "修复初始化可能执行多次", "bugfix", "medium",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar once sync.Once\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 5; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tonce.Do(func() { fmt.Println(\"init\") })\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(\"done\")\n}",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 5; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tfmt.Println(\"init\")\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(\"done\")\n}",
			"init\ndone",
			[]model.TestCase{{Output: "init"}},
			pt("once", "fmt.Println(\"init\")", "once", "medium", "初始化输出"),
			"`once.Do(...)`", "重复初始化", "多个协程各自 print 造成重复", "用 `sync.Once` 包裹初始化",
			"用 `Once.Do` 保证初始化只执行一次（此处 init 只打印一次）。"),

	// ---- 更多原语与模式 ----
	fill("go-cc-034", "sync.Cond 条件变量", "用条件变量等待某条件满足", "cond", "hard",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar mu sync.Mutex\n\tcond := sync.NewCond(&mu)\n\tready := false\n\tvar wg sync.WaitGroup\n\twg.Add(1)\n\tgo func() {\n\t\tdefer wg.Done()\n\t\tmu.Lock()\n\t\tfor !ready {\n\t\t\tcond.Wait()\n\t\t}\n\t\tfmt.Println(\"go\")\n\t\tmu.Unlock()\n\t}()\n\tmu.Lock()\n\tready = true\n\tcond.Broadcast()\n\tmu.Unlock()\n\twg.Wait()\n}",
		"go",
		[]seedPoint{pt("wait", "cond.Wait()", "cond", "hard", "等待条件"), pt("bc", "cond.Broadcast()", "cond", "hard", "唤醒所有等待者")},
		"`cond.Wait()` / `cond.Broadcast()`", "条件变量", "不满足条件则 Wait 释放锁，条件就绪后 Broadcast", "`for !ready { cond.Wait() }`",
		"`Cond.Wait` 等待并释放锁，`Signal/Broadcast` 唤醒；必须循环判断条件。"),

	fill("go-cc-035", "sync.Pool 对象池", "复用对象减少分配", "pool", "hard",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tp := &sync.Pool{New: func() any { return 0 }}\n\tx := p.Get()\n\t_ = x\n\tp.Put(42)\n\ty := p.Get()\n\tfmt.Println(y)\n}",
		"42",
		[]seedPoint{pt("pool", "sync.Pool{New:", "pool", "hard", "对象池构造")},
			"`sync.Pool{New: ...}`", "sync.Pool", "Get 取对象，Put 归还复用", "`p.Get()` / `p.Put(42)`",
			"`sync.Pool` 用于缓存与复用临时对象，降低 GC 压力。"),

	fill("go-cc-036", "barrier 屏障", "用关闭 channel 同时放行多个协程", "channel", "medium",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tstart := make(chan struct{})\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 3; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\t<-start\n\t\t\tfmt.Println(\"run\")\n\t\t}()\n\t}\n\tclose(start)\n\twg.Wait()\n}",
		"run\nrun\nrun",
		[]seedPoint{pt("start", "<-start", "channel", "medium", "等待放行信号"), pt("close", "close(start)", "channel", "medium", "同时放行")},
		"`<-start` / `close(start)`", "barrier 放行", "close 后所有接收者同时解除阻塞", "`<-start` 等待，`close(start)` 放行",
		"barrier：等待方全部阻塞在某 channel 上，close 后同时放行继续执行。"),

	fill("go-cc-037", "双向转单向", "把双向 channel 作为只读参数", "channel_direction", "medium",
		"package main\n\nimport \"fmt\"\n\nfunc consume(ch <-chan int) int {\n\treturn <-ch\n}\n\nfunc main() {\n\tch := make(chan int, 1)\n\tch <- 9\n\tfmt.Println(consume(ch))\n}",
		"9",
		[]seedPoint{pt("dir", "<-chan int", "channel_direction", "medium", "只读参数类型")},
		"`<-chan int`", "单向参数", "双向 channel 可自动转成单向参数", "`func consume(ch <-chan int)`",
		"把双向 channel 传给只读参数即限制该函数只能接收。"),

	fill("go-cc-038", "atomic 读写", "用 Store/Load 安全读写整数", "atomic", "medium",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync/atomic\"\n)\n\nfunc main() {\n\tvar n int64 = 7\n\tatomic.StoreInt64(&n, 10)\n\tfmt.Println(atomic.LoadInt64(&n))\n}",
		"10",
		[]seedPoint{pt("store", "atomic.StoreInt64(&n, 10)", "atomic", "medium", "原子写入"), pt("load", "atomic.LoadInt64(&n)", "atomic", "medium", "原子读取")},
		"`StoreInt64` / `LoadInt64`", "原子读写", "Store 写、Load 读，始终一致", "`atomic.StoreInt64(&n, 10)`",
		"用 `Store/Load` 原子读写整数，避免数据竞争。"),

	fill("go-cc-039", "懒加载单例", "用 Once 实现线程安全的懒加载", "once", "hard",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nvar (\n\tonce sync.Once\n\tcfg  string\n)\n\nfunc config() string {\n\tonce.Do(func() { cfg = \"loaded\" })\n\treturn cfg\n}\n\nfunc main() {\n\tfmt.Println(config(), config())\n}",
		"loaded loaded",
		[]seedPoint{pt("once", "once.Do", "once", "hard", "只初始化一次")},
		"`once.Do(...)`", "懒加载单例", "首次调用才初始化，且只一次", "`once.Do(func(){ cfg = \"loaded\" })`",
		"用 `sync.Once` 保证全局配置只初始化一次，后续直接复用。"),

	bugfix("go-cc-040", "并发 map 修复", "修复并发写 map 导致的 panic", "bugfix", "hard",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar mu sync.RWMutex\n\tm := map[string]int{}\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 5; i++ {\n\t\twg.Add(1)\n\t\tgo func(n int) {\n\t\t\tdefer wg.Done()\n\t\t\tmu.Lock()\n\t\t\tm[\"k\"] = n\n\t\t\tmu.Unlock()\n\t\t}(i)\n\t}\n\twg.Wait()\n\tmu.RLock()\n\tfmt.Println(len(m))\n\tmu.RUnlock()\n}",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tm := map[string]int{}\n\tvar wg sync.WaitGroup\n\tfor i := 0; i < 5; i++ {\n\t\twg.Add(1)\n\t\tgo func(n int) {\n\t\t\tdefer wg.Done()\n\t\t\tm[\"k\"] = n\n\t\t}(i)\n\t}\n\twg.Wait()\n\tfmt.Println(len(m))\n}",
		"1",
		[]model.TestCase{{Output: "1"}},
		pt("lock", "m[\"k\"] = n", "mutex", "hard", "写 map 操作"),
		"`mu.Lock(); m[\"k\"] = n; mu.Unlock()`", "并发写 map", "并发写同一 map 会 panic", "`mu.Lock()` 保护写入",
		"Go 的 map 并发写不安全，写map 要加锁（读用 RLock）。"),

	fill("go-cc-041", "ticker 定时器", "用 time.Ticker 周期性触发", "ticker", "hard",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"time\"\n)\n\nfunc main() {\n\tt := time.NewTicker(10 * time.Millisecond)\n\tgo func() {\n\t\t<-t.C\n\t\tfmt.Println(\"tick\")\n\t\tt.Stop()\n\t}()\n\ttime.Sleep(50 * time.Millisecond)\n}",
		"tick",
		[]seedPoint{pt("ticker", "time.NewTicker(10 * time.Millisecond)", "ticker", "hard", "周期定时器")},
		"`time.NewTicker`", "ticker", "`.C` 通道定期收到时间", "`<-t.C`",
		"`time.Ticker` 周期向 `.C` 通道发送、`Stop()` 停止，用于限流/心跳。"),

	fill("go-cc-042", "channel 限流令牌", "用容量 1 的 channel 串行化关键段", "channel", "hard",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n\t\"sync/atomic\"\n)\n\nfunc main() {\n\ttoken := make(chan struct{}, 1)\n\tvar wg sync.WaitGroup\n\tvar executed int64\n\tfor i := 0; i < 5; i++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\ttoken <- struct{}{}\n\t\t\tatomic.AddInt64(&executed, 1)\n\t\t\t<-token\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(executed)\n}",
		"5",
		[]seedPoint{pt("acquire", "token <- struct{}{}", "channel", "hard", "占用令牌槽"), pt("release", "<-token", "channel", "hard", "释放令牌槽")},
		"`token <- struct{}{}` / `<-token`", "Channel 限流", "容量 1 保证关键段同一时刻只有一个协程", "`token <- struct{}{}` 占用、`<-token` 释放",
		"用缓冲容量 1 的 channel 当互斥令牌：占用后执行，完事释放，共 5 次。"),

	fill("go-cc-043", "非阻塞发送", "缓冲满时走 default 不阻塞", "select", "easy",
		"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 1)\n\tch <- 1\n\tselect {\n\tcase ch <- 2:\n\t\tfmt.Println(\"sent\")\n\tdefault:\n\t\tfmt.Println(\"full\")\n\t}\n}",
		"full",
		[]seedPoint{pt("default", "default:", "select", "easy", "满则走默认")},
		"`default:`", "非阻塞发送", "缓冲满时 default 立即执行", "`case ch <- 2:` vs `default:`",
		"配合 `default` 实现非阻塞发送：满则不再等待直接走默认分支。"),

	fill("go-cc-044", "防止 goroutine 泄漏", "用 done 通道退出循环协程", "channel", "hard",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tdone := make(chan struct{})\n\tvar wg sync.WaitGroup\n\twg.Add(1)\n\tgo func() {\n\t\tdefer wg.Done()\n\t\tfor {\n\t\t\tselect {\n\t\t\tcase <-done:\n\t\t\t\treturn\n\t\t\t}\n\t\t}\n\t}()\n\tclose(done)\n\twg.Wait()\n\tfmt.Println(\"ok\")\n}",
		"ok",
		[]seedPoint{pt("done", "<-done", "channel", "hard", "收到退出信号")},
		"`case <-done: return`", "优雅退出", "done 关闭后 select 命中考量退出", "`close(done)`",
		"长期运行的协程要用 done/context 提供退出路径，避免泄漏。"),

	fill("go-cc-045", "worker 优雅关闭", "通过 done 通道关闭工作协程", "pattern", "hard",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc worker(done <-chan struct{}, wg *sync.WaitGroup) {\n\tdefer wg.Done()\n\tfor {\n\t\tselect {\n\t\tcase <-done:\n\t\t\treturn\n\t\tdefault:\n\t\t}\n\t}\n}\n\nfunc main() {\n\tdone := make(chan struct{})\n\tvar wg sync.WaitGroup\n\twg.Add(1)\n\tgo worker(done, &wg)\n\tclose(done)\n\twg.Wait()\n\tfmt.Println(\"stopped\")\n}",
		"stopped",
		[]seedPoint{pt("done", "<-done", "pattern", "hard", "关闭信号"), pt("ret", "return", "pattern", "hard", "退出 worker")},
		"`case <-done: return`", "优雅关闭", "done 关闭后 worker 退出", "`close(done)` → `wg.Wait()`",
		"worker 循环中用 select 监听 done，收到关闭信号即退出。"),

	fill("go-cc-046", "select 收集多路", "用 select 循环接收多个 channel", "select", "medium",
		"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tc1 := make(chan int, 1)\n\tc2 := make(chan int, 1)\n\tc1 <- 1\n\tc2 <- 2\n\ts := 0\n\tfor i := 0; i < 2; i++ {\n\t\tselect {\n\t\tcase v := <-c1:\n\t\t\ts += v\n\t\tcase v := <-c2:\n\t\t\ts += v\n\t\t}\n\t}\n\tfmt.Println(s)\n}",
		"3",
		[]seedPoint{pt("select", "select", "select", "medium", "多路复用")},
		"`select` 循环", "多路收集", "每次 select 取一路就绪值", "`case v := <-c1 / <-c2`",
		"用 select 循环从多个 channel 取数，两路各取一次，和 1+2=3。"),

	fill("go-cc-047", "Fan-out 分发", "一个生产者分发到多个消费者", "channel_pattern", "hard",
		"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tsrc := make(chan int, 3)\n\tgo func() {\n\t\tfor i := 1; i <= 3; i++ {\n\t\t\tsrc <- i\n\t\t}\n\t\tclose(src)\n\t}()\n\tvar mu sync.Mutex\n\ts := 0\n\tvar wg sync.WaitGroup\n\tfor w := 0; w < 2; w++ {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tfor v := range src {\n\t\t\t\tmu.Lock()\n\t\t\t\ts += v\n\t\t\t\tmu.Unlock()\n\t\t\t}\n\t\t}()\n\t}\n\twg.Wait()\n\tfmt.Println(s)\n}",
		"6",
		[]seedPoint{pt("range", "for v := range src", "channel_pattern", "hard", "消费端遍历")},
		"`for v := range src`", "Fan-out 分发", "多个消费者竞争消费同一 channel", "`close(src)` 结束分发",
		"Fan-out：一个生产者、多个消费者并行处理，总量 1+2+3=6。"),

	fill("go-cc-048", "接收 ok 提前退出", "用 ok 判断 channel 关闭", "channel", "easy",
		"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 2)\n\tch <- 1\n\tch <- 2\n\tclose(ch)\n\ts := 0\n\tfor {\n\t\tv, ok := <-ch\n\t\tif !ok {\n\t\t\tbreak\n\t\t}\n\t\ts += v\n\t}\n\tfmt.Println(s)\n}",
		"3",
		[]seedPoint{pt("ok", "ok", "channel", "easy", "通道是否关闭"), pt("break", "break", "channel", "easy", "关闭后退出")},
		"`v, ok := <-ch` / `break`", "ok 判断退出", "ok=false 表示已关闭，退出循环", "`if !ok { break }`",
		"手动循环需用 `(值, ok)` 判断 channel 是否关闭并 `break`。"),

	fill("go-cc-049", "len 统计缓冲", "接收后查看剩余元素", "channel", "easy",
		"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 3)\n\tch <- 1\n\tch <- 2\n\tch <- 3\n\t<-ch\n\t<-ch\n\tfmt.Println(len(ch))\n}",
		"1",
		[]seedPoint{pt("len", "len(ch)", "channel", "easy", "已存元素数")},
		"`len(ch)`", "channel 元素数", "发送 3 个、接收 2 个后剩 1", "`fmt.Println(len(ch))`",
		"`len(ch)` 返回缓冲中还剩的元素个数。"),

	fill("go-cc-050", "select 带 ok 接收", "在 select 中接收并判断 ok", "select", "medium",
		"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 1)\n\tch <- 7\n\tselect {\n\tcase v, ok := <-ch:\n\t\tfmt.Println(v, ok)\n\t}\n}",
		"7 true",
		[]seedPoint{pt("case", "case v, ok := <-ch:", "select", "medium", "带 ok 的接收分支")},
		"`case v, ok := <-ch:`", "select 带 ok", "在 case 里同时拿到值与 ok", "`fmt.Println(v, ok)`",
		"select 的 case 也可写成 `v, ok := <-ch`，同时判断是否成功接收。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, g := range seeds {
		ex, err := buildExercise(toGoCSeed(g))
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
