package seed

import (
	"codegym/backend/internal/model"
)

// pyBasicSeed declares one Python basic question. kind: "fill" (填空) or
// "bugfix". Every question is a complete runnable Python program judged by
// running it with python3 and comparing stdout.
type pyBasicSeed struct {
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

func toPySeed(g pyBasicSeed) seedExercise {
	exType := model.ExerciseTypeFill
	if g.kind == "bugfix" {
		exType = model.ExerciseTypeBugfix
	}
	return seedExercise{
		id: g.id, title: g.title, description: g.desc, exerciseType: exType,
		techStack: "python", weight: g.weight, category: g.category, subcategory: "基础",
		fullCode: g.code, buggyCode: g.buggy, expected: g.expected, testCases: g.testCases,
		points: g.points, hints: goHint(g.hintL1, g.hintL2, g.hintL3, g.hintL4),
		explanation: g.explanation,
	}
}

// PythonBasics returns the 50-question Python basic bank (基础).
func PythonBasics() ([]model.Exercise, error) {
	pt := func(id, text, category, weight, hint string) seedPoint {
		return seedPoint{id: id, text: text, category: category, minWeight: weight,
			weightScore: 1, hint: hint, fillType: "text"}
	}
	fill := func(id, title, desc, category, weight, code, expected string, points []seedPoint, l4, l1, l2, l3, expl string) pyBasicSeed {
		return pyBasicSeed{id: id, title: title, desc: desc, category: category, weight: weight,
			kind: "fill", code: code, expected: expected, points: points,
			hintL1: l1, hintL2: l2, hintL3: l3, hintL4: l4, explanation: expl}
	}

	seeds := []pyBasicSeed{
		fill("py-basics-001", "Hello Python", "输出问候语", "syntax", "beginner",
			"print(\"Hello, Python\")",
			"Hello, Python",
			[]seedPoint{pt("msg", "Hello, Python", "string", "beginner", "要输出的字符串")},
			"`\"Hello, Python\"`", "输出一个字符串", "用 `print(\"Hello, Python\")`", "`print(\"Hello, Python\")`",
			"Python 用 `print()` 输出，字符串用引号包裹。"),

		fill("py-basics-002", "变量与 f-string", "用 f-string 格式化字符串", "syntax", "easy",
			"name = \"Alice\"\nprint(f\"Hi, {name}!\")",
			"Hi, Alice!",
			[]seedPoint{pt("fstr", "{name}", "string", "easy", "f-string 中插入变量的写法")},
			"`{name}`", "f-string 插入变量", "`f\"Hi, {name}!\"`", "`f\"Hi, {name}!\"`",
			"f-string 以 `f` 开头，用 `{}` 插入变量或表达式。"),

		fill("py-basics-003", "幂运算", "计算 2 的 10 次方", "arithmetic", "beginner",
			"print(2 ** 10)",
			"1024",
			[]seedPoint{pt("exp", "2 ** 10", "expression", "beginner", "幂运算表达式")},
			"`2 ** 10`", "幂运算", "`**` 是幂运算符", "`print(2 ** 10)`",
			"`**` 计算幂，`2 ** 10 = 1024`；`//` 整除、`%` 取模。"),

		fill("py-basics-004", "整除与取模", "计算商和余数", "arithmetic", "beginner",
			"print(17 // 5, 17 % 5)",
			"3 2",
			[]seedPoint{pt("mod", "17 % 5", "expression", "easy", "取余数表达式")},
			"`17 % 5`", "整除与取模", "`//` 整除，`%` 取余", "`print(17 // 5, 17 % 5)`",
			"`17 // 5 = 3`（整除），`17 % 5 = 2`（余数）。"),

		fill("py-basics-005", "字符串转整数", "把字符串转成整数再运算", "conversion", "easy",
			"s = \"42\"\nn = int(s)\nprint(n + 1)",
			"43",
			[]seedPoint{pt("conv", "int(s)", "conversion", "easy", "字符串转整数的函数")},
			"`int(s)`", "类型转换", "`int()` 把字符串/浮点转整数", "`n = int(s)`",
			"`int()` 将内容为数字的字符串转换为整数。"),

		fill("py-basics-006", "格式化小数", "用 f-string 保留两位小数", "format", "easy",
			"price = 3.14159\nprint(f\"{price:.2f}\")",
			"3.14",
			[]seedPoint{pt("fmt", "{price:.2f}", "format", "medium", "保留两位小数的格式")},
			"`{price:.2f}`", "格式化小数", "`:.2f` 保留两位小数", "`print(f\"{price:.2f}\")`",
			"`:.2f` 把浮点按保留 2 位小数输出。"),

		fill("py-basics-007", "if 判断奇偶", "判断 7 是偶数还是奇数", "if", "beginner",
			"n = 7\nif n % 2 == 0:\n    print(\"even\")\nelse:\n    print(\"odd\")",
			"odd",
			[]seedPoint{pt("cond", "n % 2 == 0", "condition", "beginner", "偶数的判断条件")},
			"`n % 2 == 0`", "if 条件", "偶数条件 `n % 2 == 0`", "`if n % 2 == 0:`",
			"`if` 分支靠缩进（4 空格）划分代码块，`elif` 中间判断。"),

		fill("py-basics-008", "成绩等级", "用 if-elif-else 判断分数等级", "if", "easy",
			"score = 85\nif score >= 90:\n    print(\"A\")\nelif score >= 80:\n    print(\"B\")\nelif score >= 70:\n    print(\"C\")\nelse:\n    print(\"D\")",
			"B",
			[]seedPoint{pt("cond", "score >= 80", "condition", "easy", "B 等级的条件")},
			"`score >= 80`", "elif 分支", "85 处于 80-89 区间", "`elif score >= 80: print(\"B\")`",
			"`elif` 依次短路判断，从高到低写条件更清晰。"),

		fill("py-basics-009", "for 求和", "用 for 计算 1 到 10 的和", "for", "easy",
			"total = 0\nfor i in range(1, 11):\n    total += i\nprint(total)",
			"55",
			[]seedPoint{pt("range", "range(1, 11)", "range", "easy", "1 到 10（含头不含尾）")},
			"`range(1, 11)`", "range 范围", "`range(1, 11)` 产生 1..10", "`total += i`",
			"`range(1, 11)` 含 1 不含 11；`total += i` 累加。"),

		fill("py-basics-010", "while 循环", "用 while 计数到 5", "while", "easy",
			"n = 0\nwhile n < 5:\n    n += 1\nprint(n)",
			"5",
			[]seedPoint{pt("cond", "n < 5", "condition", "easy", "循环继续条件")},
			"`n < 5`", "while 条件", "`n < 5` 时继续循环", "`n += 1`",
			"`while` 在条件为真时反复执行；要确保条件终将变假避免死循环。"),

		fill("py-basics-011", "break 跳出", "遇到 3 就跳出循环", "control_flow", "easy",
			"s = \"\"\nfor i in range(1, 10):\n    if i > 3:\n        break\n    s += str(i)\nprint(s)",
			"123",
			[]seedPoint{pt("bk", "break", "keyword", "easy", "跳出循环的关键字")},
			"`break`", "跳出循环", "`break` 立即结束循环", "`if i > 3: break`",
			"`break` 提前终止整个循环，不再执行后续迭代。"),

		fill("py-basics-012", "continue 跳过", "跳过偶数只打印奇数", "control_flow", "easy",
			"for i in range(1, 6):\n    if i % 2 == 0:\n        continue\n    print(i)",
			"1\n3\n5",
			[]seedPoint{pt("ct", "continue", "keyword", "easy", "跳过本次循环")},
			"`continue`", "跳过本次", "`continue` 跳到下一次循环", "`if i % 2 == 0: continue`",
			"`continue` 跳过当前迭代剩余代码，进入下一次迭代。"),

		fill("py-basics-013", "列表索引", "访问列表下标元素", "list", "beginner",
			"lst = [10, 20, 30]\nprint(lst[1])",
			"20",
			[]seedPoint{pt("idx", "lst[1]", "list", "beginner", "取第 2 个元素")},
			"`lst[1]`", "列表下标", "下标从 0 开始", "`lst[1]` = 20",
			"列表下标从 0 开始，`lst[1]` 是第二个元素。"),

		fill("py-basics-014", "append 追加", "向列表追加元素", "list", "beginner",
			"lst = [1]\nlst.append(2)\nlst.append(3)\nprint(lst)",
			"[1, 2, 3]",
			[]seedPoint{pt("app", "append", "method", "beginner", "追加元素的方法")},
			"`append`", "追加元素", "`lst.append(2)` 加到末尾", "`print(lst)`",
			"`list.append` 在列表末尾追加元素。"),

		fill("py-basics-015", "列表切片", "取切片 data[1:4]", "list", "easy",
			"data = [0, 1, 2, 3, 4]\nprint(data[1:4])",
			"[1, 2, 3]",
			[]seedPoint{pt("slice", "data[1:4]", "slice", "easy", "含首不含尾的切片")},
			"`data[1:4]`", "列表切片", "`data[1:4]` 取下标 1..3", "`print(data[1:4])`",
			"切片 `[开始:结束]` 含开始不含结束；`data[-1]` 取末尾。"),

		fill("py-basics-016", "列表推导式平方", "用推导式生成平方列表", "comprehension", "easy",
			"print([x * x for x in range(5)])",
			"[0, 1, 4, 9, 16]",
			[]seedPoint{pt("expr", "x * x", "expression", "easy", "每个元素的平方")},
			"`x * x`", "列表推导式", "`[x * x for x in ...]` 逐个平方", "`[x * x for x in range(5)]`",
			"列表推导式 `[表达式 for 变量 in 序列]` 简洁生成新列表。"),

		fill("py-basics-017", "推导式筛选偶数", "用推导式 + if 筛选偶数", "comprehension", "easy",
			"print([x for x in range(10) if x % 2 == 0])",
			"[0, 2, 4, 6, 8]",
			[]seedPoint{pt("cond", "x % 2 == 0", "condition", "easy", "筛选偶数的条件")},
			"`x % 2 == 0`", "推导式筛选", "`if` 过滤元素", "`[x for x in range(10) if x % 2 == 0]`",
			"推导式后跟 `if` 可过滤，只保留满足条件的元素。"),

		fill("py-basics-018", "字符串长度", "用 len 求字符串长度", "builtins", "beginner",
			"print(len(\"python\"))",
			"6",
			[]seedPoint{pt("len", "len", "function", "beginner", "长度函数")},
			"`len`", "求长度", "`len(\"python\")` = 6", "`print(len(\"python\"))`",
			"`len()` 返回序列的长度（字符串/列表/元组等）。"),

		fill("py-basics-019", "sum 求和", "用 sum 计算列表之和", "builtins", "beginner",
			"print(sum([1, 2, 3, 4, 5]))",
			"15",
			[]seedPoint{pt("sum", "sum", "function", "beginner", "求和函数")},
			"`sum`", "求和", "`sum([1..5])` = 15", "`print(sum([1, 2, 3, 4, 5]))`",
			"`sum()` 对可迭代对象求和。"),

		fill("py-basics-020", "max 与 min", "求最大值与最小值", "builtins", "beginner",
			"print(max(3, 7, 2), min(3, 7, 2))",
			"7 2",
			[]seedPoint{pt("max", "max", "function", "beginner", "最大值函数")},
			"`max`", "取最大/最小", "`max(3,7,2)` = 7", "`print(max(3, 7, 2), min(3, 7, 2))`",
			"`max()`/`min()` 返回最大/最小值。"),

		fill("py-basics-021", "sorted 排序", "对列表排序", "builtins", "easy",
			"print(sorted([3, 1, 2]))",
			"[1, 2, 3]",
			[]seedPoint{pt("sort", "sorted", "function", "easy", "排序函数")},
			"`sorted`", "排序", "`sorted` 返回新排序列表", "`print(sorted([3, 1, 2]))`",
			"`sorted()` 返回排好序的新列表，不修改原列表。"),

		fill("py-basics-022", "元组访问", "访问元组元素", "tuple", "beginner",
			"t = (1, 2, 3)\nprint(t[1])",
			"2",
			[]seedPoint{pt("idx", "t[1]", "tuple", "beginner", "取元组第 2 个元素")},
			"`t[1]`", "元组下标", "元组不可变", "`t[1]` = 2",
			"元组 `()` 定义，不可变；按下标访问。"),

		fill("py-basics-023", "元组解包", "把元组解包到多个变量", "tuple", "easy",
			"a, b, c = (1, 2, 3)\nprint(a + c)",
			"4",
			[]seedPoint{pt("unpack", "a, b, c", "unpack", "easy", "多变量解包")},
			"`a, b, c`", "解包", "左边变量数=元组长度", "`a, b, c = (1, 2, 3)`",
			"序列解包把每个元素赋给对应变量。"),

		fill("py-basics-024", "字典读取", "按键读取字典值", "dict", "beginner",
			"d = {\"name\": \"py\"}\nprint(d[\"name\"])",
			"py",
			[]seedPoint{pt("key", "d[\"name\"]", "dict", "beginner", "按键取值")},
			"`d[\"name\"]`", "字典取值", "用键取值", "`print(d[\"name\"])`",
			"字典 `{键: 值}`，用 `d[键]` 取值；键不存在会报 KeyError。"),

		fill("py-basics-025", "字典写入", "向字典添加键值", "dict", "beginner",
			"d = {}\nd[\"x\"] = 1\nprint(d[\"x\"])",
			"1",
			[]seedPoint{pt("write", "d[\"x\"]", "dict", "easy", "写入键 x")},
			"`d[\"x\"]`", "字典写入", "`d[\"x\"] = 1` 新增键", "`print(d[\"x\"])`",
			"字典用 `d[键] = 值` 写入，键不存在则新增。"),

		fill("py-basics-026", "dict.get 默认值", "用 get 取键并给默认值", "dict", "easy",
			"d = {\"a\": 1}\nprint(d.get(\"z\", 0))",
			"0",
			[]seedPoint{pt("get", ".get(\"z\", 0)", "dict", "easy", "取键带默认值")},
			"`.get(\"z\", 0)`", "get 默认值", "键不存在返回默认 0", "`print(d.get(\"z\", 0))`",
			"`dict.get(键, 默认)` 键缺失时返回默认值，不会抛错。"),

		fill("py-basics-027", "遍历字典值", "累加字典所有值", "dict", "easy",
			"d = {\"a\": 1, \"b\": 2, \"c\": 3}\ntotal = 0\nfor v in d.values():\n    total += v\nprint(total)",
			"6",
			[]seedPoint{pt("vals", "d.values()", "dict", "easy", "取所有值")},
			"`d.values()`", "遍历值", "`d.values()` 返回所有值", "`for v in d.values()`",
			"`d.values()` / `d.keys()` / `d.items()` 遍历字典。"),

		fill("py-basics-028", "set 去重", "用集合去重并求个数", "set", "easy",
			"s = {1, 1, 2, 3}\nprint(len(s))",
			"3",
			[]seedPoint{pt("len", "len(s)", "set", "easy", "集合元素个数")},
			"`len(s)`", "集合去重", "集合自动去重", "`print(len(s))`",
			"集合 `{}` 自动去重，`len()` 得元素个数。"),

		fill("py-basics-029", "字符串大写", "用 upper 转大写", "string", "beginner",
			"print(\"go\".upper())",
			"GO",
			[]seedPoint{pt("upper", ".upper()", "method", "beginner", "大写方法")},
			"`.upper()`", "转大写", "`\"go\".upper()` → GO", "`print(\"go\".upper())`",
			"`str.upper()` 返回全大写；`lower()` 全小写。"),

		fill("py-basics-030", "strip 去空白", "去掉字符串首尾空格", "string", "easy",
			"print(\"  hi  \".strip())",
			"hi",
			[]seedPoint{pt("strip", ".strip()", "method", "easy", "去除首尾空白")},
			"`.strip()`", "去空白", "`\"  hi  \".strip()` → hi", "`print(\"  hi  \".strip())`",
			"`str.strip()` 去掉首尾空白，`lstrip`/`rstrip` 只去一边。"),

		fill("py-basics-031", "split 与 join", "切分再拼接字符串", "string", "easy",
			"words = \"a,b,c\".split(\",\")\nprint(\"-\".join(words))",
			"a-b-c",
			[]seedPoint{pt("split", ".split(\",\")", "method", "easy", "按逗号切分")},
			"`.split(\",\")`", "切分与拼接", "`split` 变列表，`join` 反拼接", "`\"-\".join(words)`",
			"`split()` 按分隔符切开，`join()` 用分隔符合并为字符串。"),

		fill("py-basics-032", "replace 替换", "替换字符串中的字符", "string", "easy",
			"print(\"a-b-c\".replace(\"-\", \"+\"))",
			"a+b+c",
			[]seedPoint{pt("rep", "replace(\"-\", \"+\")", "method", "easy", "替换函数")},
			"`replace(\"-\", \"+\")`", "字符串替换", "把所有 - 换成 +", "`print(\"a-b-c\".replace(\"-\", \"+\"))`",
			"`str.replace(旧, 新)` 把所有匹配子串替换。"),

		fill("py-basics-033", "find 查找", "找子串位置", "string", "easy",
			"print(\"hello\".find(\"ll\"))",
			"2",
			[]seedPoint{pt("find", "find(\"ll\")", "method", "easy", "查找子串下标")},
			"`find(\"ll\")`", "查找下标", "找不到返回 -1", "`print(\"hello\".find(\"ll\"))`",
			"`str.find(子串)` 返回首次出现的下标，找不到返回 -1。"),

		fill("py-basics-034", "startswith 判断", "判断前缀", "string", "easy",
			"print(\"python\".startswith(\"py\"))",
			"True",
			[]seedPoint{pt("starts", "startswith(\"py\")", "method", "easy", "前缀判断")},
			"`startswith(\"py\")`", "判断前缀", "返回布尔值", "`print(\"python\".startswith(\"py\"))`",
			"`startswith()`/`endswith()` 判断前缀/后缀，返回 True/False。"),

		fill("py-basics-035", "enumerate 遍历", "遍历并带下标", "iteration", "medium",
			"names = [\"a\", \"b\"]\nfor i, n in enumerate(names):\n    print(i, n)",
			"0 a\n1 b",
			[]seedPoint{pt("enum", "enumerate(names)", "function", "medium", "带下标的遍历")},
			"`enumerate(names)`", "带下标遍历", "同时拿到下标和值", "`for i, n in enumerate(names)`",
			"`enumerate()` 同时产出下标与元素，遍历时免去手动计数。"),

		fill("py-basics-036", "zip 配对", "把两个列表配对", "builtins", "medium",
			"print(list(zip([1, 2], [3, 4])))",
			"[(1, 3), (2, 4)]",
			[]seedPoint{pt("zip", "zip", "function", "medium", "配对函数")},
			"`zip`", "序列配对", "`zip` 把对应元素组成元组", "`print(list(zip([1, 2], [3, 4])))`",
			"`zip()` 把多个序列按位置配对；`list()` 转成列表查看。"),

		fill("py-basics-037", "map 转换", "把数字列表转成字符串", "functional", "medium",
			"print(list(map(str, [1, 2, 3])))",
			"['1', '2', '3']",
			[]seedPoint{pt("map", "map", "function", "medium", "批量转换函数")},
			"`map`", "批量映射", "`map(str, ...)` 逐项转字符串", "`print(list(map(str, [1, 2, 3])))`",
			"`map()` 把函数应用到序列每个元素；需 `list()` 取出结果。"),

		fill("py-basics-038", "lambda 平方", "用 lambda 定义匿名函数", "functional", "medium",
			"f = lambda x: x ** 2\nprint(f(5))",
			"25",
			[]seedPoint{pt("lambda", "lambda x: x ** 2", "lambda", "medium", "匿名函数")},
			"`lambda x: x ** 2`", "匿名函数", "`lambda 参数: 表达式`", "`f(5)` = 25",
			"`lambda` 定义一次性匿名小函数，结果就是返回值。"),

		fill("py-basics-039", "filter 筛选", "用 filter 筛选偶数", "functional", "medium",
			"print(list(filter(lambda x: x % 2 == 0, range(10))))",
			"[0, 2, 4, 6, 8]",
			[]seedPoint{pt("filter", "filter", "function", "medium", "筛选函数")},
			"`filter`", "筛选元素", "保留 lambda 为真的元素", "`print(list(filter(...)))`",
			"`filter()` 保留使函数返回真的元素，与 `lambda` 常搭配。"),

		fill("py-basics-040", "函数定义", "定义并调用 add 函数", "function", "easy",
			"def add(a, b):\n    return a + b\nprint(add(2, 3))",
			"5",
			[]seedPoint{pt("ret", "return a + b", "function", "easy", "返回值")},
			"`return a + b`", "函数定义", "`def` 定义，`return` 返回", "`add(2, 3)`",
			"`def 名(参数):` 定义函数，`return` 返回计算结果。"),

		fill("py-basics-041", "默认参数", "给函数参数设默认值", "function", "easy",
			"def greet(name=\"world\"):\n    return f\"hi {name}\"\nprint(greet())",
			"hi world",
			[]seedPoint{pt("default", "name=\"world\"", "function", "easy", "默认参数值")},
			"`name=\"world\"`", "默认参数", "不传参时用默认值", "`greet()`",
			"参数可设默认值，调用时不传则使用默认。"),

		fill("py-basics-042", "多返回值", "函数返回总和与个数", "function", "medium",
			"def stats(nums):\n    return sum(nums), len(nums)\ns, c = stats([1, 2, 3])\nprint(s, c)",
			"6 3",
			[]seedPoint{pt("ret", "sum(nums), len(nums)", "function", "medium", "返回多个值")},
			"`sum(nums), len(nums)`", "多返回值", "用逗号返回多个值", "`s, c = stats([1, 2, 3])`",
			"Python 函数可用逗号一次返回多个值，接收时解包。"),

		fill("py-basics-043", "*args 变长参数", "接受任意数量参数", "function", "medium",
			"def total(*args):\n    return sum(args)\nprint(total(1, 2, 3))",
			"6",
			[]seedPoint{pt("args", "*args", "function", "medium", "变长位置参数")},
			"`*args`", "可变参数", "`*args` 收集所有位置参数", "`total(1, 2, 3)`",
			"`*args` 收集任意数量的位置参数为元组。"),

		fill("py-basics-044", "递归阶乘", "用递归计算阶乘", "recursion", "medium",
			"def fact(n):\n    return 1 if n <= 1 else n * fact(n - 1)\nprint(fact(5))",
			"120",
			[]seedPoint{pt("rec", "n * fact(n - 1)", "recursion", "medium", "递推式")},
			"`n * fact(n - 1)`", "递归", "递归 = 基线 + 递推", "`fact(5)` = 120",
			"`fact(n) = n * fact(n-1)`，基线 `n<=1` 返回 1。"),

		fill("py-basics-045", "递归斐波那契", "用递归求第 10 个斐波那契数", "recursion", "medium",
			"def fib(n):\n    return n if n < 2 else fib(n - 1) + fib(n - 2)\nprint(fib(10))",
			"55",
			[]seedPoint{pt("rec", "fib(n - 1) + fib(n - 2)", "recursion", "medium", "递推式")},
			"`fib(n - 1) + fib(n - 2)`", "递归斐波那契", "前两项之和", "`fib(10)` = 55",
			"`fib(n) = fib(n-1) + fib(n-2)`，基线 `n<2` 返回 n。"),

		fill("py-basics-046", "try/except 捕获", "捕获转换异常", "exception", "medium",
			"try:\n    print(int(\"abc\"))\nexcept ValueError:\n    print(\"bad\")",
			"bad",
			[]seedPoint{pt("conv", "int(\"abc\")", "exception", "medium", "会抛异常的操作")},
			"`int(\"abc\")`", "异常处理", "转换失败抛 ValueError", "`except ValueError: print(\"bad\")`",
			"`try/except` 捕获异常；`int(\"abc\")` 无法转换抛出 ValueError。"),

		fill("py-basics-047", "除零异常", "捕获 ZeroDivisionError", "exception", "medium",
			"try:\n    print(1 / 0)\nexcept ZeroDivisionError:\n    print(\"no\")",
			"no",
			[]seedPoint{pt("exc", "ZeroDivisionError", "exception", "medium", "除零异常类型")},
			"`ZeroDivisionError`", "异常类型", "除零抛 ZeroDivisionError", "`except ZeroDivisionError`",
			"除以 0 抛出 `ZeroDivisionError`，用对应异常类型捕获。"),

		fill("py-basics-048", "math 开方", "用 math.sqrt 求平方根", "module", "easy",
			"import math\nprint(math.sqrt(16))",
			"4.0",
			[]seedPoint{pt("sqrt", "math.sqrt", "module", "easy", "开方函数")},
			"`math.sqrt`", "导入模块", "`import math` 后使用", "`math.sqrt(16)` = 4.0",
			"`import math` 导入标准库，`math.sqrt` 开平方。"),

		fill("py-basics-049", "自定义类", "定义类与方法", "class", "medium",
			"class Dog:\n    def __init__(self, name):\n        self.name = name\n    def bark(self):\n        return self.name + \"!\"\nd = Dog(\"rex\")\nprint(d.bark())",
			"rex!",
			[]seedPoint{pt("ret", "self.name + \"!\"", "class", "medium", "方法返回值")},
			"`self.name + \"!\"`", "类与方法", "`__init__` 构造，`self` 指实例", "`d.bark()`",
			"`class` 定义类，`__init__` 初始化，方法首参 `self` 指向实例。"),

		fill("py-basics-050", "f-string 表达式", "在 f-string 中直接计算", "format", "easy",
			"a = 3\nb = 4\nprint(f\"{a}+{b}={a + b}\")",
			"3+4=7",
			[]seedPoint{pt("expr", "{a + b}", "format", "easy", "f-string 内计算")},
			"`{a + b}`", "f-string 表达式", "`{}` 里可放任意表达式", "`f\"{a}+{b}={a + b}\"`",
			"f-string 的 `{}` 内可直接写表达式并求值插入。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, g := range seeds {
		ex, err := buildExercise(toPySeed(g))
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
