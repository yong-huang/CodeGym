package seed

import (
	"codegym/backend/internal/model"
)

// goBasicSeed declares one Go basic question. kind: "fill" (填空, real
// execution), "bugfix" (real execution) or "command" (answer-match).
type goBasicSeed struct {
	id, title, desc, category, weight string
	kind                               string
	code                               string
	buggy                              string
	expected                           string
	testCases                          []model.TestCase
	points                             []seedPoint
	hintL1, hintL2, hintL3, hintL4     string
	explanation                        string
}

// goHint builds a 4-level hint ladder, revealing the answer(s) at L4.
func goHint(l1, l2, l3, l4 string) model.SpotterHints {
	return model.SpotterHints{
		Level1: "💡 " + l1,
		Level2: "📖 " + l2,
		Level3: "👀 " + l3,
		Level4: "📝 正确答案：" + l4,
	}
}

func toGoSeed(g goBasicSeed) seedExercise {
	exType := model.ExerciseTypeFill
	if g.kind == "bugfix" {
		exType = model.ExerciseTypeBugfix
	} else if g.kind == "command" {
		exType = model.ExerciseTypeCommand
	}
	points := g.points
	if exType == model.ExerciseTypeCommand {
		points = addVerbPoint(points, g.code)
	} else if exType == model.ExerciseTypeFill {
		points = enrichGoPoints(points, g.code, 5)
	}
	return seedExercise{
		id:           g.id,
		title:        g.title,
		description:  g.desc,
		exerciseType: exType,
		techStack:    "go",
		weight:       g.weight,
		category:     g.category,
		fullCode:     g.code,
		buggyCode:    g.buggy,
		expected:     g.expected,
		testCases:    g.testCases,
		points:       points,
		hints:        goHint(g.hintL1, g.hintL2, g.hintL3, g.hintL4),
		explanation:  g.explanation,
	}
}

// GoBasics returns the 50-question Go basic bank (基础题).
func GoBasics() ([]model.Exercise, error) {
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
		fill("go-basics-001", "Hello Go", "输出问候语", "syntax", "beginner",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, Go\")\n}",
			"Hello, Go",
			[]seedPoint{pt("msg", "\"Hello, Go\"", "string", "beginner", "要输出的字符串")},
			"`\"Hello, Go\"`", "输出一个字符串", "用 `fmt.Println(\"Hello, Go\")`", "`fmt.Println(\"Hello, Go\")`",
			"`fmt.Println` 打印并换行；字符串用双引号包裹。"),

		fill("go-basics-002", "变量声明", "用 := 声明两个变量并打印", "variables", "beginner",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tname := \"gopher\"\n\tage := 5\n\tfmt.Println(name, age)\n}",
			"gopher 5",
			[]seedPoint{pt("name", "\"gopher\"", "string", "beginner", "字符串变量值"), pt("age", "5", "number", "beginner", "整数变量值")},
			"`\"gopher\"` / `5`", "用 := 声明的两个变量", "`name := \"gopher\"`，`age := 5`", "`fmt.Println(name, age)`",
			"`:=` 声明并初始化变量，类型由右侧推断；`var` 可显式声明。"),

		fill("go-basics-003", "常量 const", "声明并打印常量", "constants", "beginner",
			"package main\n\nimport \"fmt\"\n\nconst pi = 3.14159\n\nfunc main() {\n\tfmt.Println(pi)\n}",
			"3.14159",
			[]seedPoint{pt("val", "3.14159", "constant", "beginner", "圆周率常量值")},
			"`3.14159`", "常量的值", "`const pi = 3.14159`", "`fmt.Println(pi)`",
			"`const` 定义编译期常量，运行期不能修改。"),

		fill("go-basics-004", "if-else 判断奇偶", "判断 7 的奇偶性", "if", "beginner",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tn := 7\n\tif n%2 == 0 {\n\t\tfmt.Println(\"even\")\n\t} else {\n\t\tfmt.Println(\"odd\")\n\t}\n}",
			"odd",
			[]seedPoint{pt("cond", "n%2 == 0", "condition", "beginner", "偶数的判断条件")},
			"`n%2 == 0`", "偶数的取模条件", "`n%2 == 0` 为真则是偶数", "`if n%2 == 0 { } else { }`",
			"Go 的 if/else 条件不需要括号，但必须用花括号包裹分支。"),

		fill("go-basics-005", "for 循环 1 到 5", "用 for 打印 1..5", "for", "beginner",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfor i := 1; i <= 5; i++ {\n\t\tfmt.Println(i)\n\t}\n}",
			"1\n2\n3\n4\n5",
			[]seedPoint{pt("cond", "i <= 5", "condition", "beginner", "循环继续条件")},
			"`i <= 5`", "循环的终止条件", "`i <= 5` 时继续", "`for i := 1; i <= 5; i++ { }`",
			"Go 只有 `for` 一种循环；`初始;条件;步进` 三段式写法类似其它语言。"),

		fill("go-basics-006", "switch 匹配分支", "用 switch 判断星期", "switch", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tday := \"mon\"\n\tswitch day {\n\tcase \"mon\":\n\t\tfmt.Println(\"start\")\n\tcase \"fri\":\n\t\tfmt.Println(\"tgif\")\n\tdefault:\n\t\tfmt.Println(\"other\")\n\t}\n}",
			"start",
			[]seedPoint{pt("case", "case \"mon\":", "case", "easy", "匹配 day==\"mon\" 的分支")},
			"`case \"mon\":`", "匹配字符串的分支写法", "`case \"mon\":` 匹配 day", "`switch day { case \"mon\": ... }`",
			"Go 的 switch 无需 break，匹配到分支后自动退出。"),

		fill("go-basics-007", "数组求和", "用 range 遍历数组求和", "array", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tnums := [4]int{10, 20, 30, 40}\n\tsum := 0\n\tfor _, n := range nums {\n\t\tsum += n\n\t}\n\tfmt.Println(sum)\n}",
			"100",
			[]seedPoint{pt("acc", "sum += n", "statement", "easy", "累加每个元素")},
			"`sum += n`", "累加当前元素", "`for _, n := range nums` 遍历", "`sum += n`",
			"`range` 遍历返回索引和值，用 `_` 忽略索引；对数组、切片、map、字符串都适用。"),

		fill("go-basics-008", "切片追加 append", "向切片追加元素", "slice", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tnums := []int{1, 2}\n\tnums = append(nums, 3, 4)\n\tfmt.Println(nums)\n}",
			"[1 2 3 4]",
			[]seedPoint{pt("fn", "append", "function", "beginner", "追加元素的内置函数")},
			"`append`", "切片追加的内置函数", "`append(nums, 3, 4)`", "`nums = append(nums, 3, 4)`",
			"`append` 追加元素，容量不够时自动扩容；必须把返回值赋回变量。"),

		fill("go-basics-009", "切片切片", "取 a[1:3] 生成新切片", "slice", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\ta := []int{1, 2, 3, 4, 5}\n\tb := a[1:3]\n\tfmt.Println(b)\n}",
			"[2 3]",
			[]seedPoint{pt("slice", "a[1:3]", "slice", "easy", "截取下标 1 到 2")},
			"`a[1:3]`", "截取下标 1..2", "`a[1:3]` 含头不含尾", "`b := a[1:3]`",
			"切片 `a[low:high]` 含 low 不含 high，这里是下标 1、2。" ),

		fill("go-basics-010", "make 创建切片", "用 make 创建长度为 3 的切片", "make", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\ts := make([]int, 3)\n\tfmt.Println(len(s))\n}",
			"3",
			[]seedPoint{pt("make", "make([]int, 3)", "function", "easy", "创建容量为 3 的切片")},
			"`make([]int, 3)`", "创建切片的函数", "`make([]int, 3)` 长度 3", "`s := make([]int, 3)`",
			"`make` 用于创建切片、map、channel；第二个参数是长度。"),

		fill("go-basics-011", "map 写入", "向 map 写入并读取", "map", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tcounts := map[string]int{\"a\": 1, \"b\": 2}\n\tcounts[\"c\"] = 3\n\tfmt.Println(counts[\"c\"])\n}",
			"3",
			[]seedPoint{pt("write", "counts[\"c\"] = 3", "statement", "easy", "写入键 c")},
			"`counts[\"c\"] = 3`", "写入 map 键值", "`counts[\"c\"] = 3`", "`fmt.Println(counts[\"c\"])`",
			"map 用 `m[key] = value` 写入，`m[key]` 读取；读取不存在的键返回零值。"),

		fill("go-basics-012", "字符串拼接", "拼接字符串并查看长度", "string", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tgreet := \"Hello, \" + \"Go\"\n\tfmt.Println(greet, len(greet))\n}",
			"Hello, Go 9",
			[]seedPoint{pt("part", "\"Hello, \"", "string", "easy", "拼接的一部分")},
			"`\"Hello, \"`", "要拼接的字符串", "`\"Hello, \" + \"Go\"`", "`len(greet)`",
			"字符串用 `+` 拼接；`len` 返回字节数。" ),

		fill("go-basics-013", "函数返回", "定义 add 函数并调用", "function", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc add(a, b int) int {\n\treturn a + b\n}\n\nfunc main() {\n\tfmt.Println(add(2, 3))\n}",
			"5",
			[]seedPoint{pt("ret", "return a + b", "statement", "easy", "返回两数之和")},
			"`return a + b`", "返回表达式", "`func add(a, b int) int`", "`return a + b`",
			"函数签名 `func 名(参数) 返回类型`，`return` 返回值。"),

		fill("go-basics-014", "多返回值", "函数返回商和余数", "function", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc div(a, b int) (int, int) {\n\treturn a / b, a % b\n}\n\nfunc main() {\n\tq, r := div(10, 3)\n\tfmt.Println(q, r)\n}",
			"3 1",
			[]seedPoint{pt("sig", "(int, int)", "signature", "medium", "两个返回值的类型列表")},
			"`(int, int)`", "多返回值的类型", "`(int, int)` 声明两个 int 返回", "`q, r := div(10, 3)`",
			"Go 支持多返回值，常用于 `返回值, error` 的惯用法。"),

		fill("go-basics-015", "具名返回", "用具名返回值求和", "function", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc sum(nums []int) (total int) {\n\tfor _, n := range nums {\n\t\ttotal += n\n\t}\n\treturn\n}\n\nfunc main() {\n\tfmt.Println(sum([]int{1, 2, 3}))\n}",
			"6",
			[]seedPoint{pt("sig", "(total int)", "signature", "medium", "具名返回值的声明")},
			"`(total int)`", "具名返回值写法", "`(total int)` 定义变量 total", "裸 `return` 返回具名变量",
			"具名返回值在函数内自动声明，裸 `return` 返回它们。"),

		fill("go-basics-016", "可变参数", "用 ...int 接受任意数量参数取最大值", "variadic", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc max(nums ...int) int {\n\tm := nums[0]\n\tfor _, n := range nums {\n\t\tif n > m {\n\t\t\tm = n\n\t\t}\n\t}\n\treturn m\n}\n\nfunc main() {\n\tfmt.Println(max(3, 7, 2))\n}",
			"7",
			[]seedPoint{pt("sig", "...int", "variadic", "medium", "可变参数的写法")},
			"`...int`", "可变参数写法", "`nums ...int` 收集成切片", "`max(3, 7, 2)`",
			"`...T` 声明可变参数，函数内表现为类型 T 的切片。"),

		fill("go-basics-017", "defer 延迟执行", "打印后再执行延迟语句", "defer", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tdefer fmt.Println(\"done\")\n\tfmt.Println(\"start\")\n}",
			"start\ndone",
			[]seedPoint{pt("defer", "defer", "keyword", "easy", "延迟执行关键字")},
			"`defer`", "延迟执行的关键字", "`defer` 在函数返回时执行", "`defer fmt.Println(\"done\")`",
			"`defer` 延迟到函数返回前执行，常用于关闭资源，多个 defer 按后进先出。"),

		fill("go-basics-018", "闭包计数器", "闭包捕获并累计变量", "closure", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc counter() func() int {\n\tn := 0\n\treturn func() int {\n\t\tn++\n\t\treturn n\n\t}\n}\n\nfunc main() {\n\tc := counter()\n\tfmt.Println(c(), c(), c())\n}",
			"1 2 3",
			[]seedPoint{pt("ret", "return func() int", "closure", "medium", "返回一个函数字面量")},
			"`return func() int`", "返回匿名函数", "闭包捕获外层变量 n", "`func() int { n++; return n }`",
			"闭包可以捕获并修改外层局部变量，形成带状态函数。"),

		fill("go-basics-019", "字符串转整数", "用 strconv.Atoi 解析字符串", "strconv", "easy",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"strconv\"\n)\n\nfunc main() {\n\tn, err := strconv.Atoi(\"123\")\n\tif err != nil {\n\t\tfmt.Println(\"bad\")\n\t} else {\n\t\tfmt.Println(n)\n\t}\n}",
			"123",
			[]seedPoint{pt("fn", "strconv.Atoi", "function", "easy", "字符串转整数的函数")},
			"`strconv.Atoi`", "字符串转整数", "`strconv.Atoi(\"123\")` 返回 (int, error)", "`n, err := strconv.Atoi(\"123\")`",
			"`Atoi` 把字符串解析为 int，第二个返回值 error 需检查。"),

		fill("go-basics-020", "返回错误", "函数返回自定义 error", "error", "medium",
			"package main\n\nimport (\n\t\"errors\"\n\t\"fmt\"\n)\n\nfunc half(x int) (int, error) {\n\tif x%2 != 0 {\n\t\treturn 0, errors.New(\"odd\")\n\t}\n\treturn x / 2, nil\n}\n\nfunc main() {\n\tv, err := half(8)\n\tif err != nil {\n\t\tfmt.Println(err)\n\t} else {\n\t\tfmt.Println(v)\n\t}\n}",
			"4",
			[]seedPoint{pt("err", "errors.New(\"odd\")", "error", "medium", "构造错误对象")},
			"`errors.New(\"odd\")`", "构造一个错误", "`errors.New` 返回 error 值", "`return 0, errors.New(\"odd\")`",
			"`errors.New` 创建一个简单的 error，惯用法是返回 `值, nil` 表示成功。"),

		fill("go-basics-021", "结构体", "定义并初始化结构体", "struct", "easy",
			"package main\n\nimport \"fmt\"\n\ntype User struct {\n\tName string\n\tAge  int\n}\n\nfunc main() {\n\tu := User{Name: \"gopher\", Age: 3}\n\tfmt.Println(u.Name, u.Age)\n}",
			"gopher 3",
			[]seedPoint{pt("init", "User{Name: \"gopher\", Age: 3}", "struct", "easy", "结构体字面量")},
			"`User{Name: \"gopher\", Age: 3}`", "结构体初始化", "`User{Name:..., Age:...}`", "`u.Name` 访问字段",
			"`type X struct` 定义复合类型；`X{字段: 值}` 构造实例。"),

		fill("go-basics-022", "结构体方法", "定义圆面积方法", "method", "medium",
			"package main\n\nimport \"fmt\"\n\ntype Circle struct {\n\tRadius float64\n}\n\nfunc (c Circle) Area() float64 {\n\treturn 3.14 * c.Radius * c.Radius\n}\n\nfunc main() {\n\tc := Circle{Radius: 2}\n\tfmt.Println(c.Area())\n}",
			"12.56",
			[]seedPoint{pt("m", "func (c Circle) Area()", "method", "medium", "方法的接收者与名")},
			"`func (c Circle) Area()`", "方法定义格式", "`(c Circle)` 是接收者", "`c.Area()` 调用",
			"方法 = 带接收者的函数；接收者 `c Circle` 使函数绑定到类型。"),

		fill("go-basics-023", "指针参数", "用指针在函数内修改变量", "pointer", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc setVal(x *int) {\n\t*x = 42\n}\n\nfunc main() {\n\tn := 0\n\tsetVal(&n)\n\tfmt.Println(n)\n}",
			"42",
			[]seedPoint{pt("ptype", "*int", "pointer", "medium", "指向 int 的指针类型"), pt("addr", "&n", "pointer", "medium", "取变量地址")},
			"`*int` / `&n`", "指针类型与取地址", "`*int` 声明指针，`&n` 取地址", "`*x = 42` 解引用赋值",
			"`*` 解引用，`&` 取地址；通过指针可在函数内修改调用方变量。"),

		fill("go-basics-024", "接口", "用接口类型保存正方形并调用方法", "interface", "medium",
			"package main\n\nimport \"fmt\"\n\ntype Shape interface {\n\tArea() float64\n}\n\ntype Square struct {\n\tSide float64\n}\n\nfunc (s Square) Area() float64 {\n\treturn s.Side * s.Side\n}\n\nfunc main() {\n\tvar s Shape = Square{Side: 4}\n\tfmt.Println(s.Area())\n}",
			"16",
			[]seedPoint{pt("var", "var s Shape", "interface", "medium", "声明接口变量")},
			"`var s Shape`", "声明接口类型变量", "`Square` 自动满足 `Shape` 接口", "`var s Shape = Square{Side: 4}`",
			"接口定义方法集合，类型实现全部方法即满足接口（隐式实现）。"),

		fill("go-basics-025", "指针接收者", "用指针接收者修改结构体字段", "method", "medium",
			"package main\n\nimport \"fmt\"\n\ntype Counter struct {\n\tn int\n}\n\nfunc (c *Counter) Inc() {\n\tc.n++\n}\n\nfunc main() {\n\tc := &Counter{}\n\tc.Inc()\n\tfmt.Println(c.n)\n}",
			"1",
			[]seedPoint{pt("m", "func (c *Counter) Inc()", "method", "medium", "指针接收者的方法")},
			"`func (c *Counter) Inc()`", "指针接收者", "`(c *Counter)` 可修改字段", "`c := &Counter{}`",
			"指针接收者方法能修改结构体字段；`&Counter{}` 取实例地址。"),

		fill("go-basics-026", "Goroutine 并发求和", "用 goroutine + 互斥锁并发求和", "concurrency", "hard",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tvar mu sync.Mutex\n\tsum := 0\n\tfor i := 1; i <= 10; i++ {\n\t\twg.Add(1)\n\t\tgo func(n int) {\n\t\t\tdefer wg.Done()\n\t\t\tmu.Lock()\n\t\t\tsum += n\n\t\t\tmu.Unlock()\n\t\t}(i)\n\t}\n\twg.Wait()\n\tfmt.Println(sum)\n}",
			"55",
			[]seedPoint{pt("go", "go func", "keyword", "easy", "启动 goroutine 的关键字"), pt("wait", "wg.Wait()", "statement", "easy", "等待所有 goroutine 完成")},
			"`go func` / `wg.Wait()`", "启动协程与等待", "`go` 关键字启动协程，`WaitGroup` 等待", "`wg.Wait()` 阻塞直到计数归零",
			"`go` 启动并发 goroutine；`WaitGroup.Wait` 等待；Mutex 保证并发写互斥。"),

		fill("go-basics-027", "Channel 发送接收", "向有缓冲 channel 发送并接收", "channel", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 1)\n\tch <- 7\n\tfmt.Println(<-ch)\n}",
			"7",
			[]seedPoint{pt("recv", "<-ch", "channel", "easy", "从 channel 接收值")},
			"`<-ch`", "从 channel 取值", "`ch <- 7` 发送，`<-ch` 接收", "`fmt.Println(<-ch)`",
			"`chan` 类型通信；`ch <- v` 发送，`v := <-ch` 接收；带缓冲容量 1 不阻塞发送。"),

		fill("go-basics-028", "Channel 通信", "goroutine 间用 channel 传值", "channel", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int)\n\tgo func() { ch <- 1 }()\n\tv := <-ch\n\tfmt.Println(v)\n}",
			"1",
			[]seedPoint{pt("ch", "make(chan int)", "channel", "medium", "创建无缓冲 channel")},
			"`make(chan int)`", "创建无缓冲 channel", "无缓冲发送会阻塞直到接收", "`go func() { ch <- 1 }()`",
			"无缓冲 channel 发送和接收同步阻塞；用 goroutine 发、主程序收。"),

		fill("go-basics-029", "select 多路选择", "用 select 从有值的 channel 接收", "select", "hard",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\ta := make(chan int, 1)\n\tb := make(chan int, 1)\n\tb <- 2\n\tselect {\n\tcase x := <-a:\n\t\tfmt.Println(\"from a\", x)\n\tcase x := <-b:\n\t\tfmt.Println(\"from b\", x)\n\t}\n}",
			"from b 2",
			[]seedPoint{pt("case", "case x := <-b:", "case", "hard", "接收 channel b 的分支")},
			"`case x := <-b:`", "select 的接收分支", "`b` 有值则走该分支", "`select { case x := <-b: ... }`",
			"`select` 同时监听多个 channel，任一就绪即执行对应分支。"),

		fill("go-basics-030", "遍历 channel", "用 range 累加 channel 直至关闭", "channel", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tch := make(chan int, 3)\n\tch <- 1\n\tch <- 2\n\tch <- 3\n\tclose(ch)\n\tsum := 0\n\tfor v := range ch {\n\t\tsum += v\n\t}\n\tfmt.Println(sum)\n}",
			"6",
			[]seedPoint{pt("close", "close(ch)", "statement", "medium", "关闭 channel 的函数")},
			"`close(ch)`", "关闭 channel", "`range ch` 在 channel 关闭后结束", "`for v := range ch`",
			"`close` 关闭 channel；`range ch` 遍历直到关闭。"),

		fill("go-basics-031", "类型转换", "把 int 显式转为 float64", "conversion", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tvar i int = 42\n\tvar f float64 = float64(i)\n\tfmt.Printf(\"%.1f\\n\", f)\n}",
			"42.0",
			[]seedPoint{pt("conv", "float64(i)", "conversion", "easy", "int 到 float64 的转换")},
			"`float64(i)`", "数值类型转换", "不同类型不能直接赋值", "`var f float64 = float64(i)`",
			"Go 是强类型，数值类型转换用 `T(value)` 显式写法。"),

		fill("go-basics-032", "fmt 格式动词", "用 %d %x %b 输出十进制/十六进制/二进制", "fmt", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tn := 255\n\tfmt.Printf(\"%d %x %b\\n\", n, n, n)\n}",
			"255 ff 11111111",
			[]seedPoint{pt("verb", "%d %x %b", "format", "easy", "指定多个格式动词")},
			"`%d %x %b`", "三种进制格式", "`%d` 十进制 `%x` 十六进制 `%b` 二进制", "`fmt.Printf(\"%d %x %b\\n\", n, n, n)`",
			"`%d` 十进制、`%x` 十六进制、`%b` 二进制；`Printf` 按格式输出。"),

		fill("go-basics-033", "按 rune 计数", "区分字节数与字符数", "string", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\ts := \"你好Go\"\n\tfmt.Println(len(s), len([]rune(s)))\n}",
			"8 4",
			[]seedPoint{pt("rune", "[]rune(s)", "conversion", "medium", "转成 rune 切片")},
			"`[]rune(s)`", "按字符计数", "`len(s)` 是字节数，`[]rune(s)` 是字符数", "`fmt.Println(len(s), len([]rune(s)))`",
			"UTF-8 中汉字占 3 字节；`len` 返回字节数，`len([]rune(s))` 返回字符数。"),

		fill("go-basics-034", "strings 去除空白", "用 TrimSpace 去掉首尾空格", "strings", "easy",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc main() {\n\ts := \"  hello  \"\n\tfmt.Println(strings.TrimSpace(s))\n}",
			"hello",
			[]seedPoint{pt("fn", "strings.TrimSpace", "function", "easy", "去除首尾空白的函数")},
			"`strings.TrimSpace`", "去除首尾空白", "`strings.TrimSpace` 返回去空格字符串", "`fmt.Println(strings.TrimSpace(s))`",
			"`strings.TrimSpace` 去掉首尾空白字符。"),

		fill("go-basics-035", "strings 包含判断", "判断字符串是否含子串", "strings", "easy",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc main() {\n\tfmt.Println(strings.Contains(\"hello world\", \"world\"))\n}",
			"true",
			[]seedPoint{pt("fn", "strings.Contains", "function", "easy", "判断子串包含的函数")},
			"`strings.Contains`", "判断包含", "`strings.Contains(s, sub)` 返回 bool", "`fmt.Println(strings.Contains(\"hello world\", \"world\"))`",
			"`strings.Contains` 判断是否包含子串，返回布尔值。"),

		fill("go-basics-036", "math 最大值", "用 math.Max 求最大值", "math", "easy",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"math\"\n)\n\nfunc main() {\n\tfmt.Println(math.Max(3, 7))\n}",
			"7",
			[]seedPoint{pt("fn", "math.Max", "function", "easy", "浮点取最大值的函数")},
			"`math.Max`", "取最大值", "`math.Max` 返回较大者", "`fmt.Println(math.Max(3, 7))`",
			"`math.Max` 对 float64 取最大值；int 无内置 max，可用泛型或比较。"),

		fill("go-basics-037", "sort 排序", "对整数切片升序排序", "sort", "easy",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"sort\"\n)\n\nfunc main() {\n\tnums := []int{3, 1, 2}\n\tsort.Ints(nums)\n\tfmt.Println(nums)\n}",
			"[1 2 3]",
			[]seedPoint{pt("fn", "sort.Ints", "function", "easy", "整数切片排序函数")},
			"`sort.Ints`", "整数切片排序", "`sort.Ints` 原地升序", "`sort.Ints(nums)`",
			"`sort.Ints` 对 []int 原地升序排序。"),

		fill("go-basics-038", "copy 复制切片", "用 copy 复制切片内容", "slice", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\ta := []int{1, 2, 3}\n\tb := make([]int, len(a))\n\tcopy(b, a)\n\tfmt.Println(b)\n}",
			"[1 2 3]",
			[]seedPoint{pt("fn", "copy", "function", "easy", "复制切片的内置函数")},
			"`copy`", "复制切片", "`copy(dst, src)` 拷贝", "`copy(b, a)`",
			"`copy` 把源切片内容拷到目标切片；目标需先分配长度。" ),

		fill("go-basics-039", "无限循环与 break", "用 for{} + break 跳出", "control_flow", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tn := 0\n\tfor {\n\t\tn++\n\t\tif n > 3 {\n\t\t\tbreak\n\t\t}\n\t}\n\tfmt.Println(n)\n}",
			"4",
			[]seedPoint{pt("bk", "break", "keyword", "easy", "跳出循环的关键字")},
			"`break`", "跳出循环", "`for { }` 无限循环，`break` 跳出", "`if n > 3 { break }`",
			"`for {}` 是无限循环，用 `break` 提前退出。"),

		fill("go-basics-040", "continue 跳过", "跳过偶数累加奇数", "control_flow", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tsum := 0\n\tfor i := 1; i <= 5; i++ {\n\t\tif i%2 == 0 {\n\t\t\tcontinue\n\t\t}\n\t\tsum += i\n\t}\n\tfmt.Println(sum)\n}",
			"9",
			[]seedPoint{pt("ct", "continue", "keyword", "easy", "跳过本次循环")},
			"`continue`", "跳过本次循环", "偶数时 `continue` 不参与累加", "`if i%2 == 0 { continue }`",
			"`continue` 跳过当前迭代剩下的语句直接进入下一次。"),

		fill("go-basics-041", "遍历 map 值", "累加 map 所有值", "map", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tm := map[string]int{\"a\": 1, \"b\": 2}\n\ttotal := 0\n\tfor _, v := range m {\n\t\ttotal += v\n\t}\n\tfmt.Println(total)\n}",
			"3",
			[]seedPoint{pt("range", "for _, v := range", "range", "easy", "遍历 map 的语法")},
			"`for _, v := range`", "遍历 map 值", "`range m` 迭代键值", "`for _, v := range m`",
			"`range` map 每次迭代返回 key, value；无序遍历的和是确定的。"),

		fill("go-basics-042", "iota 自增常量", "用 iota 生成连续常量", "constants", "medium",
			"package main\n\nimport \"fmt\"\n\nconst (\n\tA = iota\n\tB\n\tC\n)\n\nfunc main() {\n\tfmt.Println(A, B, C)\n}",
			"0 1 2",
			[]seedPoint{pt("iota", "iota", "keyword", "medium", "自增常量生成器")},
			"`iota`", "自增常量生成器", "`iota` 在 const 块内从 0 自增", "`A = iota; B; C`",
			"`iota` 是行的自增索引，常用于定义枚举常量。"),

		fill("go-basics-043", "switch 无表达式", "用无表达式的 switch 实现多条件", "switch", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc classify(n int) string {\n\tswitch {\n\tcase n < 0:\n\t\treturn \"negative\"\n\tcase n == 0:\n\t\treturn \"zero\"\n\tdefault:\n\t\treturn \"positive\"\n\t}\n}\n\nfunc main() {\n\tfmt.Println(classify(-5))\n}",
			"negative",
			[]seedPoint{pt("cond", "n < 0", "condition", "easy", "负数条件的表达式")},
			"`n < 0`", "无表达式的 switch 条件", "`switch {}` 各 case 是布尔表达式", "`case n < 0:`",
			"`switch` 不带表达式时每个 case 写一个条件，等价 if-elseif。"),

		fill("go-basics-044", "类型别名", "定义基于 float64 的类型别名", "types", "easy",
			"package main\n\nimport \"fmt\"\n\ntype Celsius float64\n\nfunc main() {\n\ttemp := Celsius(36.6)\n\tfmt.Println(temp)\n}",
			"36.6",
			[]seedPoint{pt("alias", "type Celsius float64", "type", "easy", "类型别名声明")},
			"`type Celsius float64`", "基于现有类型定义新类型", "`Celsius(36.6)` 转换", "`type Celsius float64`",
			"`type 名称 底层类型` 定义命名类型，方便表达领域含义。"),

		fill("go-basics-045", "泛型函数", "定义支持多数值类型的 Max 泛型函数", "generics", "hard",
			"package main\n\nimport \"fmt\"\n\nfunc Max[T int | float64](a, b T) T {\n\tif a > b {\n\t\treturn a\n\t}\n\treturn b\n}\n\nfunc main() {\n\tfmt.Println(Max(3, 7))\n\tfmt.Println(Max(2.5, 1.5))\n}",
			"7\n2.5",
			[]seedPoint{pt("constraint", "[T int | float64]", "generics", "hard", "泛型约束")},
			"`[T int | float64]`", "泛型类型参数与约束", "`T` 限定为 int 或 float64", "`Max[T int | float64]`",
			"Go 1.18+ 支持泛型，`[T 约束]` 声明类型参数，约束限定允许的类型。"),

		bugfix("go-basics-046", "闰年判断 Bug", "修复闰年判断只检查被 4 整除的 bug", "bugfix", "medium",
			"package main\n\nimport \"fmt\"\n\nfunc leap(y int) bool {\n\treturn y%4 == 0 && (y%100 != 0 || y%400 == 0)\n}\n\nfunc main() {\n\tfmt.Println(leap(2020))\n\tfmt.Println(leap(1900))\n}",
			"package main\n\nimport \"fmt\"\n\nfunc leap(y int) bool {\n\treturn y%4 == 0\n}\n\nfunc main() {\n\tfmt.Println(leap(2020))\n\tfmt.Println(leap(1900))\n}",
			"true\nfalse",
			[]model.TestCase{{Output: "true"}, {Output: "false"}},
			pt("cond", "return y%4 == 0", "condition", "medium", "只判断整除 4 是错的，需排除整百年"),
			"`return y%4 == 0 && (y%100 != 0 || y%400 == 0)`",
			"只看 e%4 漏判世纪年", "1900 被 100 整除但非 400 倍数", "加 `(y%100 != 0 || y%400 == 0)`",
			"闰年 = 被 4 整除且（不被 100 整除或 被 400 整除）；1900 不是闰年。"),

		bugfix("go-basics-047", "平均值 Bug", "修复未除以元素个数的 bug", "bugfix", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc avg(nums []int) int {\n\tsum := 0\n\tfor _, n := range nums {\n\t\tsum += n\n\t}\n\treturn sum / len(nums)\n}\n\nfunc main() {\n\tfmt.Println(avg([]int{1, 2, 3, 4}))\n\tfmt.Println(avg([]int{10, 20}))\n}",
			"package main\n\nimport \"fmt\"\n\nfunc avg(nums []int) int {\n\tsum := 0\n\tfor _, n := range nums {\n\t\tsum += n\n\t}\n\treturn sum\n}\n\nfunc main() {\n\tfmt.Println(avg([]int{1, 2, 3, 4}))\n\tfmt.Println(avg([]int{10, 20}))\n}",
			"2\n15",
			[]model.TestCase{{Output: "2"}, {Output: "15"}},
			pt("ret", "return sum", "statement", "easy", "返回的是平均应除以个数"),
			"`return sum / len(nums)`", "返回了总和而非平均值", "`len(nums)` 是元素个数", "`return sum / len(nums)`",
			"平均值 = 总和 / 元素个数，需除以 `len(nums)`。"),

		bugfix("go-basics-048", "交换变量 Bug", "修复两变量未真正交换的 bug", "bugfix", "easy",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\ta, b := 1, 2\n\ta, b = b, a\n\tfmt.Println(a, b)\n}",
			"package main\n\nimport \"fmt\"\n\nfunc main() {\n\ta, b := 1, 2\n\ta = b\n\tb = a\n\tfmt.Println(a, b)\n}",
			"2 1",
			[]model.TestCase{{Output: "2 1"}},
			pt("swap", "a = b", "statement", "beginner", "逐个赋值到不了交换的目的"),
			"`a, b = b, a`", "单步赋值会丢失一个值", "先 a=b 后 b=a 会让 a、b 都是原 b", "用 `a, b = b, a` 简洁交换",
			"多变量同时赋值可直接交换；逐个赋值需要临时变量。"),

		fill("go-basics-049", "元音计数", "统计字符串中元音个数", "string_loop", "medium",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc main() {\n\ts := \"go is fun\"\n\tcount := 0\n\tfor _, r := range s {\n\t\tif strings.ContainsRune(\"aeiou\", r) {\n\t\t\tcount++\n\t\t}\n\t}\n\tfmt.Println(count)\n}",
			"3",
			[]seedPoint{pt("rune", "strings.ContainsRune", "function", "medium", "判断 rune 是否在字符串中")},
			"`strings.ContainsRune`", "按字符判断包含", "`strings.ContainsRune(s, r)` 判断是否含该字符", "`if strings.ContainsRune(\"aeiou\", r)`",
			"`range` 遍历字符串产生 rune；`ContainsRune` 判断是否在元音集合中。"),

		fill("go-basics-050", "切分字符串", "用 strings.Split 切分并计数", "strings", "easy",
			"package main\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc main() {\n\tparts := strings.Split(\"a,b,c\", \",\")\n\tfmt.Println(len(parts))\n}",
			"3",
			[]seedPoint{pt("fn", "strings.Split", "function", "easy", "按分隔符切分的函数")},
			"`strings.Split`", "按分隔符切分", "`strings.Split(s, \",\")` 返回切片", "`fmt.Println(len(parts))`",
			"`strings.Split(s, sep)` 按分隔符切分字符串，返回切片。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, g := range seeds {
		ex, err := buildExercise(toGoSeed(g))
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
