package seed

import (
	"fmt"
	"strings"
	"time"

	"codegym/backend/internal/model"
)

// seedPoint declares a fill point by searching for its text in full_code,
// so span offsets are always correct even after the code text is edited.
type seedPoint struct {
	id                string
	text              string
	category          string
	minWeight         string
	weightScore       float64
	hint              string
	fillType          string
	options           []string
	acceptableAnswers []string
	alwaysFill        bool
	occurrence        int // 0-based occurrence of `text` to use
}

type seedExercise struct {
	id           string
	title        string
	description  string
	exerciseType string
	techStack    string
	weight       string
	category     string
	subcategory  string
	fullCode     string
	buggyCode    string
	bugDesc      string
	bugLoc       string
	expected     string
	validation   string
	testCases    []model.TestCase
	points       []seedPoint
	hints        model.SpotterHints
	explanation  string
}

func buildExercise(s seedExercise) (model.Exercise, error) {
	ex := model.Exercise{
		ID:             s.id,
		Title:          s.title,
		Description:    s.description,
		ExerciseType:   s.exerciseType,
		TechStack:      s.techStack,
		Weight:         s.weight,
		Category:       s.category,
		Subcategory:    s.subcategory,
		FullCode:       s.fullCode,
		BuggyCode:      s.buggyCode,
		BugDescription: s.bugDesc,
		BugLocation:    s.bugLoc,
		ExpectedOutput: s.expected,
		ValidationCode: s.validation,
		TestCases:      s.testCases,
		SpotterHints:   s.hints,
		TimeLimitSecs:  30,
		Explanation:    s.explanation,
		Enabled:        true,
		CreatedBy:      "seed",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		FillPoints: model.FillPointsConfig{
			MinFills: 1,
			MaxFills: len(s.points),
		},
	}

	code := s.fullCode
	if s.exerciseType == model.ExerciseTypeBugfix {
		code = s.buggyCode
	}
	for _, p := range s.points {
		start, ok := occurrenceIndex(code, p.text, p.occurrence)
		if !ok {
			return ex, fmt.Errorf("seed %s: text %q occurrence %d not found in code", s.id, p.text, p.occurrence)
		}
		line := strings.Count(code[:start], "\n") + 1
		ex.FillPoints.Points = append(ex.FillPoints.Points, model.FillPoint{
			ID:                p.id,
			Location:          model.FillLocation{Type: "span", Line: line, SpanStart: start, SpanEnd: start + len(p.text)},
			OriginalText:      p.text,
			AcceptableAnswers: p.acceptableAnswers,
			Category:          p.category,
			MinWeight:         p.minWeight,
			WeightScore:       p.weightScore,
			AlwaysFill:        p.alwaysFill,
			Hint:              p.hint,
			FillType:          p.fillType,
			Options:           p.options,
		})
	}
	// large config YAML gets extra fill points so higher weights fill more
	if s.exerciseType == model.ExerciseTypeConfig {
		extra := enrichYamlPoints(s.points, code, 6)
		if len(extra) > len(s.points) {
			for i := len(s.points); i < len(extra); i++ {
				p := extra[i]
				start, ok := occurrenceIndex(code, p.text, 0)
				if ok {
					ex.FillPoints.Points = append(ex.FillPoints.Points, model.FillPoint{
						ID: p.id, Location: model.FillLocation{Type: "span", SpanStart: start, SpanEnd: start + len(p.text)},
						OriginalText: p.text, Category: p.category, MinWeight: p.minWeight,
						WeightScore: p.weightScore, Hint: p.hint, FillType: p.fillType,
					})
				}
			}
		}
	}
	ensureBeginnerPoint(ex.FillPoints.Points)
	return ex, nil
}

// ensureBeginnerPoint guarantees every exercise has at least one training
// point fillable at the beginner level. Otherwise low weight levels would have
// no blank to fill (a broken/empty exercise).
func ensureBeginnerPoint(points []model.FillPoint) {
	ranks := map[string]int{"beginner": 1, "easy": 2, "medium": 3, "hard": 4, "expert": 5}
	hasFillableAtBeginner := false
	lowest := 99
	lowIdx := -1
	for i, p := range points {
		if p.MinWeight == "" || ranks[p.MinWeight] <= 1 {
			hasFillableAtBeginner = true
		}
		if r := ranks[p.MinWeight]; r > 0 && r < lowest {
			lowest, lowIdx = r, i
		}
	}
	if !hasFillableAtBeginner && lowIdx >= 0 {
		points[lowIdx].MinWeight = "beginner"
	}
}

// occurrenceIndex returns the start of the nth occurrence of sub in s.
func occurrenceIndex(s, sub string, n int) (int, bool) {
	idx := -1
	for i := 0; i <= n; i++ {
		next := strings.Index(s[idx+1:], sub)
		if next < 0 {
			return 0, false
		}
		idx += next + 1
	}
	return idx, true
}

// addVerbPoint gives a single-point command exercise a second fill on its
// subcommand, so higher weight levels visibly fill more of the command
// (e.g. beginner fills only the flag, harder levels also fill the verb).
func addVerbPoint(points []seedPoint, code string) []seedPoint {
	if len(points) < 1 {
		return points
	}
	fields := strings.Fields(strings.TrimSpace(code))
	if len(fields) < 3 {
		return points // single-verb commands (e.g. `kubectl version`) stay as-is
	}
	verb := fields[1]
	for _, p := range points {
		if p.text == verb {
			return points
		}
	}
	return append(points, seedPoint{
		id:         "verb_" + sanitizeID(verb),
		text:       verb,
		category:   "command_verb",
		minWeight:  "easy",
		weightScore: 0.5,
		hint:       "这句命令的子命令",
		fillType:   "text",
	})
}

// goTokenHints lists salient Go concurrency tokens to auto-add as fill points.
var goTokenHints = []struct {
	pattern, category, hint string
}{
	{"go func", "goroutine", "用 go 关键字启动协程"},
	{"wg.Add(1)", "waitgroup", "启动前计数器 +1"},
	{"wg.Wait()", "waitgroup", "等待所有协程完成"},
	{"defer wg.Done()", "waitgroup", "协程完成时 -1"},
	{"mu.Lock()", "mutex", "进入临界区加锁"},
	{"mu.Unlock()", "mutex", "退出临界区解锁"},
	{"mu.RLock()", "rwmutex", "读操作加读锁"},
	{"once.Do", "once", "只执行一次的初始化"},
	{"atomic.AddInt64", "atomic", "原子自增"},
	{"atomic.StoreInt64", "atomic", "原子写入"},
	{"atomic.LoadInt64", "atomic", "原子读取"},
	{"atomic.CompareAndSwapInt64", "atomic", "原子比较交换"},
	{"cond.Wait()", "cond", "等待条件满足"},
	{"cond.Broadcast()", "cond", "唤醒所有等待者"},
	{"time.NewTicker", "ticker", "周期定时器"},
	{"time.After(", "timeout", "超时通道"},
	{"ctx.Done()", "context", "取消/超时通知"},
	{"cancel()", "context", "触发取消"},
	{"<-done", "channel", "接收退出信号"},
	{"close(done)", "channel", "关闭即广播放行"},
	{"close(ch)", "channel", "关闭 channel 结束接收"},
	{"make(chan", "channel", "创建 channel"},
	{"ch <- ", "channel", "向 channel 发送"},
	{"<-ch", "channel", "从 channel 接收"},
	{"select {", "select", "多路复用选择"},
}

// enrichGoPoints scans a Go program and appends concurrency fill points for
// salient tokens (go/waitgroup/mutex/chan/context…), each with escalating
// difficulty, so higher weight levels visibly fill more. Skips texts already
// covered and caps additions.
func enrichGoPoints(points []seedPoint, code string, maxAdded int) []seedPoint {
	hasText := func(t string) bool {
		for _, p := range points {
			if p.text == t {
				return true
			}
		}
		return false
	}
	levels := []string{"easy", "easy", "medium", "medium", "hard", "hard", "expert"}
	li := 0
	added := 0
	for _, t := range goTokenHints {
		if added >= maxAdded {
			break
		}
		if !strings.Contains(code, t.pattern) || hasText(t.pattern) {
			continue
		}
		lvl := levels[li]
		if li < len(levels)-1 {
			li++
		}
		points = append(points, seedPoint{
			id: "auto_" + sanitizeID(t.pattern), text: t.pattern, category: t.category,
			minWeight: lvl, weightScore: 0.7, hint: t.hint, fillType: "text",
		})
		added++
	}
	return points
}

// sanitizeID makes an arbitrary token usable as a fill point id.
func sanitizeID(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "x"
	}
	return b.String()
}

// yamlFillHints maps config field keys to a short hint for auto-added points.
var yamlFillHints = map[string]string{
	"replicas":      "副本数",
	"image":         "容器镜像",
	"containerPort": "容器端口",
	"targetPort":    "后端目标端口",
	"nodePort":      "节点端口",
	"port":          "服务/探针端口",
	"storage":       "存储容量",
	"cpu":           "CPU 配额",
	"memory":        "内存配额",
	"host":          "域名/主机名",
	"uri":           "源/仓库地址",
	"path":          "路径",
	"medium":        "介质/类型",
	"termination":   "终止方式",
	"strategy":      "构建策略",
	"reclaimPolicy": "回收策略",
	"provisioner":   "卷供应器",
	"app":           "标签 app 的值",
}

// enrichYamlPoints inspects a config YAML and appends additional fill points
// for salient scalar fields (replicas, image, ports, storage, cpu…), each with
// an escalating difficulty, so higher weight levels visibly fill more.
// Small configs (a few fields) are left with their existing points; only
// larger YAML blocks get extra fill points (capped at maxAdded).
func enrichYamlPoints(points []seedPoint, code string, maxAdded int) []seedPoint {
	// count meaningful YAML lines; small snippets stay simple (1-2 points)
	meaningful := 0
	var lines []string
	for _, ln := range strings.Split(code, "\n") {
		t := strings.TrimSpace(ln)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		meaningful++
		lines = append(lines, t)
	}
	if meaningful < 6 {
		return points
	}

	hasText := func(t string) bool {
		for _, p := range points {
			if p.text == t {
				return true
			}
		}
		return false
	}
	levels := []string{"easy", "easy", "medium", "medium", "hard", "hard", "expert"}
	li := 0
	added := 0
	for _, t := range lines {
		if added >= maxAdded {
			break
		}
		if strings.HasPrefix(t, "- ") {
			continue
		}
		idx := strings.Index(t, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(t[:idx])
		hint, ok := yamlFillHints[key]
		if !ok || hint == "" {
			continue
		}
		val := strings.Trim(strings.TrimSpace(t[idx+1:]), `'"+ `)
		if val == "" || strings.ContainsAny(val, "{[]}/\\") || len(val) > 40 {
			continue
		}
		if hasText(val) {
			continue
		}
		lvl := levels[li]
		if li < len(levels)-1 {
			li++
		}
		points = append(points, seedPoint{
			id: "auto_" + key + "_" + sanitizeID(val), text: val, category: "config",
			minWeight: lvl, weightScore: 0.7, hint: "补充「" + hint + "」", fillType: "text",
		})
		added++
	}
	return points
}

// CategoryPair is a (tech_stack, subcategory) default managed category.
type CategoryPair struct {
	Tech  string
	Sub   string
}

// CategoryPlan returns the default set of managed subcategories. Empty ones
// (高级/实战/web) are just manageable containers until questions are added.
func CategoryPlan() []CategoryPair {
	return []CategoryPair{
		{Tech: "python", Sub: "基础"},
		{Tech: "javascript", Sub: "基础"},
		{Tech: "go", Sub: "基础"},
		{Tech: "go", Sub: "高级"},
		{Tech: "go", Sub: "实战"},
		{Tech: "go", Sub: "web"},
		{Tech: "kubernetes", Sub: "基础"},
		{Tech: "kubernetes", Sub: "资源配置"},
		{Tech: "kubernetes", Sub: "资源配置"},
		{Tech: "kubernetes", Sub: "高级"},
		{Tech: "kubernetes", Sub: "实战"},
		{Tech: "kubernetes", Sub: "web"},
		{Tech: "openshift", Sub: "基础"},
		{Tech: "openshift", Sub: "对比"},
		{Tech: "openshift", Sub: "高级"},
		{Tech: "openshift", Sub: "实战"},
		{Tech: "openshift", Sub: "web"},
		{Tech: "docker", Sub: "基础"},
		{Tech: "docker", Sub: "高级"},
		{Tech: "docker", Sub: "实战"},
		{Tech: "docker", Sub: "web"},
	}
}

// All returns every seeded exercise, ready for Upsert.
func All() ([]model.Exercise, error) {
	seedExercises := []seedExercise{
		{
			id:          "py-lc-001",
			title:       "列表推导式 - 偶数平方",
			description: "将 numbers 列表中的偶数筛选出来并计算平方，生成新列表",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "python",
			weight:      model.DefaultWeightLevel,
			category:    "list_comprehension",
			fullCode:    "numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]\nresult = [x**2 for x in numbers if x % 2 == 0]\nprint(result)",
			expected:    "[4, 16, 36, 64, 100]",
			points: []seedPoint{
				{id: "expr", text: "x**2", category: "expression", minWeight: "beginner", weightScore: 0.6,
					hint: "计算每个元素的平方，可以使用幂运算符或乘法", fillType: "text"},
				{id: "cond", text: "x % 2 == 0", category: "condition", minWeight: "medium", weightScore: 0.8,
					hint: "偶数的数学特征是什么？能被什么整除？", fillType: "text"},
				{id: "iter", text: "numbers", category: "iterator", minWeight: "easy", weightScore: 0.4,
					hint: "对哪个可迭代对象做推导？", fillType: "text", occurrence: 1},
			},
			hints: model.SpotterHints{
				Level1: "💡 想一想，偶数的数学特征是什么？",
				Level2: "📖 偶数能被 2 整除，用取模运算符 % 来判断",
				Level3: "👀 条件表达式应该是 `x % 2 == 0` 这样的形式",
				Level4: "📝 正确答案：`x % 2 == 0`",
			},
			explanation: "列表推导式 `[expr for item in iterable if condition]`。`x**2` 计算平方，`x % 2 == 0` 筛选偶数，两者都满足时保留并平方。",
		},
		{
			id:          "py-palindrome-002",
			title:       "回文判断函数",
			description: "实现 is_palindrome 函数，判断字符串是否正反相同",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "python",
			weight:      model.DefaultWeightLevel,
			category:    "function",
			fullCode:    "def is_palindrome(s: str) -> bool:\n    return s == s[::-1]\n\nprint(is_palindrome('racecar'))\nprint(is_palindrome('hello'))",
			expected:    "True\nFalse",
			points: []seedPoint{
				{id: "return_expr", text: "s[::-1]", category: "expression", minWeight: "beginner", weightScore: 0.5,
					hint: "想想字符串反转的方法（切片）", fillType: "text"},
				{id: "return_type", text: "-> bool", category: "type_annotation", minWeight: "easy", weightScore: 0.3,
					hint: "返回值是什么类型？布尔值", fillType: "text"},
				{id: "compare", text: "s == s[::-1]", category: "condition", minWeight: "medium", weightScore: 0.8,
					hint: "正序和反转后相等即为回文", fillType: "text", occurrence: 0},
				{id: "call_arg", text: "'hello'", category: "argument", minWeight: "hard", weightScore: 0.6,
					hint: "第二个测试用例的反例参数", fillType: "text", occurrence: 0},
			},
			hints: model.SpotterHints{
				Level1: "💡 判断回文只需比较原串与反转后的串",
				Level2: "📖 切片 `s[::-1]` 可以反转字符串",
				Level3: "👀 `return s == s[::-1]`",
				Level4: "📝 正确答案：`return s == s[::-1]`",
			},
			explanation: "切片 `s[::-1]` 返回反转后的字符串。`s == s[::-1]` 在正反相同（回文）时返回 True。",
		},
		{
			id:          "py-factorial-bug-003",
			title:       "阶乘 Bug 修复",
			description: "find and fix the off-by-one bug in factorial",
			exerciseType: model.ExerciseTypeBugfix,
			techStack:   "python",
			weight:      "hard",
			category:    "bugfix",
			fullCode:    "def factorial(n):\n    result = 1\n    for i in range(1, n + 1):\n        result *= i\n    return result",
			buggyCode:   "def factorial(n):\n    result = 1\n    for i in range(n):\n        result *= i\n    return result",
			bugDesc:     "n=5 时输出 0，预期 120。range(n) 从 0 开始，result 乘以 0 后永远为 0",
			bugLoc:      "for i in range(n)",
			expected:    "120\n6",
			validation:  "\nprint(factorial(5))\nprint(factorial(3))",
			testCases:   []model.TestCase{{Output: "120"}, {Output: "6"}},
			points: []seedPoint{
				{id: "loop", text: "for i in range(n)", category: "loop", minWeight: "beginner", weightScore: 1,
					hint: "range(n) 从 0 开始，为什么结果会是 0？", fillType: "code"},
			},
			hints: model.SpotterHints{
				Level1: "💡 仔细看循环的起始值，乘以 0 会发生什么？",
				Level2: "📖 range(n) 生成 [0..n-1]，第一次循环 result *= 0 把结果清成 0",
				Level3: "👀 循环应从 1 开始到 n 结束",
				Level4: "📝 正确答案：`for i in range(1, n + 1):`",
			},
			explanation: "off-by-one：range(n) 从 0 开始，第一次迭代 `result *= 0` 使结果恒为 0。改为 `range(1, n + 1)` 即可。",
		},
		{
			id:          "py-fizzbuzz-004",
			title:       "FizzBuzz 函数",
			description: "实现 FizzBuzz：3 的倍数输出 Fizz，5 的倍数输出 Buzz，同时是 3 和 5 的倍数输出 FizzBuzz",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "python",
			weight:      "medium",
			category:    "control_flow",
			fullCode:    "def fizzbuzz(n):\n    if n % 15 == 0:\n        return 'FizzBuzz'\n    if n % 3 == 0:\n        return 'Fizz'\n    if n % 5 == 0:\n        return 'Buzz'\n    return str(n)\n\nprint(fizzbuzz(15))\nprint(fizzbuzz(9))\nprint(fizzbuzz(10))\nprint(fizzbuzz(7))",
			expected:    "FizzBuzz\nFizz\nBuzz\n7",
			points: []seedPoint{
				{id: "fizzbuzz", text: "n % 15 == 0", category: "condition", minWeight: "beginner", weightScore: 0.8,
					hint: "同时是 3 和 5 的倍数的最小公倍数是多少？", fillType: "text"},
				{id: "fizz", text: "n % 3 == 0", category: "condition", minWeight: "easy", weightScore: 0.5,
					hint: "3 的倍数如何判断？", fillType: "text"},
				{id: "buzz", text: "n % 5 == 0", category: "condition", minWeight: "easy", weightScore: 0.5,
					hint: "5 的倍数如何判断？", fillType: "text"},
				{id: "fallback", text: "str(n)", category: "expression", minWeight: "medium", weightScore: 0.4,
					hint: "都不是倍数时返回数字本身，注意类型", fillType: "text"},
			},
			hints: model.SpotterHints{
				Level1: "💡 三个条件都有取模运算",
				Level2: "📖 n % 15 == 0 表示能同时被 3 和 5 整除",
				Level3: "👀 依次判断 15、3、5 的倍数",
				Level4: "📝 正确答案：`n % 15 == 0` / `n % 3 == 0` / `n % 5 == 0` / `str(n)`",
			},
			explanation: "先判断是否同时是 3 和 5 的倍数（15 的倍数），再分别判断 3 和 5，最后返回原数字的字符串形式。",
		},
		{
			id:          "kubectl-getpods-005",
			title:       "kubectl 查看所有命名空间的 Pod",
			description: "补全命令，查看所有命名空间的 Pod",
			exerciseType: model.ExerciseTypeCommand,
			techStack:   "kubernetes",
			weight:      "easy",
			category:    "command",
			fullCode:    "# 查看所有命名空间的 Pod\nkubectl get pods -A",
			expected:    "",
			points: []seedPoint{
				{id: "flag", text: "-A", category: "command_flag", minWeight: "beginner", weightScore: 1,
					hint: "可以用短参数 -A 或长参数 --all-namespaces", fillType: "select",
					options: []string{"-A", "--all-namespaces", "-n default", "--namespace all"},
					acceptableAnswers: []string{"--all-namespaces"}},
			},
			hints: model.SpotterHints{
				Level1: "💡 所有命名空间的关键词是什么？",
				Level2: "📖 `--all-namespaces` 或短参数 `-A`",
				Level3: "👀 `kubectl get pods -A`",
				Level4: "📝 正确答案：`-A`（等价 `--all-namespaces`）",
			},
			explanation: "`-A` 是 `--all-namespaces` 的短参数，列出所有命名空间中的 Pod。",
		},
		{
			id:          "k8s-nginx-pod-006",
			title:       "Nginx Pod 配置",
			description: "补全 Kubernetes Pod YAML 中的关键字段",
			exerciseType: model.ExerciseTypeConfig,
			techStack:   "kubernetes",
			weight:      "medium",
			category:    "manifest",
			fullCode:    "apiVersion: v1\nkind: Pod\nmetadata:\n  name: nginx-pod\nspec:\n  containers:\n  - name: nginx\n    image: nginx:1.25\n    ports:\n    - containerPort: 80",
			expected:    "",
			points: []seedPoint{
				{id: "image", text: "nginx:1.25", category: "image", minWeight: "beginner", weightScore: 0.7,
					hint: "Nginx 镜像及版本标签", fillType: "text"},
				{id: "port", text: "80", category: "port", minWeight: "easy", weightScore: 0.6,
					hint: "Nginx 默认监听端口", fillType: "text"},
				{id: "name", text: "nginx-pod", category: "metadata", minWeight: "medium", weightScore: 0.5,
					hint: "Pod 的名称字段", fillType: "text"},
			},
			hints: model.SpotterHints{
				Level1: "💡 镜像与端口分别填写什么？",
				Level2: "📖 镜像 `nginx:1.25`，Nginx 默认 80 端口",
				Level3: "👀 `image: nginx:1.25` / `containerPort: 80`",
				Level4: "📝 正确答案：`nginx:1.25` 与 `80`",
			},
			explanation: "Pod 资源要求镜像名称（含标签）与容器端口，`nginx:1.25` 对应官方镜像，Nginx 默认监听 80。",
		},
		{
			id:          "js-even-squares-007",
			title:       "JS 数组处理 - 偶数平方",
			description: "用 filter 和 map 生成偶数的平方数组",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "javascript",
			weight:      model.DefaultWeightLevel,
			category:    "array_methods",
			fullCode:    "const numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];\nconst result = numbers.filter(x => x % 2 === 0).map(x => x ** 2);\nconsole.log(JSON.stringify(result));",
			expected:    "[4,16,36,64,100]",
			points: []seedPoint{
				{id: "filter_cond", text: "x % 2 === 0", category: "condition", minWeight: "medium", weightScore: 0.7,
					hint: "筛选偶数的条件（严格相等）", fillType: "text"},
				{id: "map_fn", text: "x ** 2", category: "expression", minWeight: "beginner", weightScore: 0.5,
					hint: "计算平方的箭头函数体", fillType: "text"},
			},
			hints: model.SpotterHints{
				Level1: "💡 filter 筛选，map 变换",
				Level2: "📖 偶数条件 `x % 2 === 0`，平方 `x ** 2`",
				Level3: "👀 `.filter(x => x % 2 === 0).map(x => x ** 2)`",
				Level4: "📝 正确答案：`x % 2 === 0` 与 `x ** 2`",
			},
			explanation: "`filter` 保留满足条件的元素，`map` 对每个元素做变换，链式调用即可完成筛选并平方。",
		},
		// ---------- Go 题库 ----------
		{
			id:          "go-even-squares-101",
			title:       "Geven 偶数平方过滤",
			description: "用 for + 取模筛选偶数并计算平方，收集到新的切片",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "go",
			weight:      model.DefaultWeightLevel,
			category:    "slice",
			fullCode:    "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tnumbers := []int{1, 2, 3, 4, 5, 6, 7, 8}\n\tresult := []int{}\n\tfor _, n := range numbers {\n\t\tif n%2 == 0 {\n\t\t\tresult = append(result, n*n)\n\t\t}\n\t}\n\tfmt.Println(result)\n}",
			expected:    "[4 16 36 64]",
			points: []seedPoint{
				{id: "cond", text: "n%2 == 0", category: "condition", minWeight: "easy", weightScore: 0.7,
					hint: "用取模判断偶数的条件是什么？", fillType: "text"},
				{id: "square", text: "n*n", category: "expression", minWeight: "beginner", weightScore: 0.6,
					hint: "计算 n 的平方（乘法表达式）", fillType: "text"},
				{id: "append", text: "append", category: "function", minWeight: "easy", weightScore: 0.5,
					hint: "向切片末尾追加元素的 Go 内置函数", fillType: "text"},
			},
			hints: model.SpotterHints{
				Level1: "💡 偶数能被 2 整除，平方用乘法",
				Level2: "📖 判断 `n%2 == 0`，计算 `n*n`，追加 `append`",
				Level3: "👀 `n%2 == 0` → `result = append(result, n*n)`",
				Level4: "📝 正确答案：`n%2 == 0` / `n*n` / `append`",
			},
			explanation: "`n%2 == 0` 判断偶数，`n*n` 计算平方，`append` 向切片追加元素。",
		},
		{
			id:          "go-palindrome-102",
			title:       "Goure 回文判断",
			description: "实现 isPalindrome，比较字符串首尾字符判断回文",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "go",
			weight:      model.DefaultWeightLevel,
			category:    "string",
			fullCode:    "package main\n\nimport \"fmt\"\n\nfunc isPalindrome(s string) bool {\n\trunes := []rune(s)\n\tfor i := 0; i < len(runes); i++ {\n\t\tif runes[i] != runes[len(runes)-1-i] {\n\t\t\treturn false\n\t\t}\n\t}\n\treturn true\n}\n\nfunc main() {\n\tfmt.Println(isPalindrome(\"racecar\"))\n\tfmt.Println(isPalindrome(\"hello\"))\n}",
			expected:    "true\nfalse",
			points: []seedPoint{
				{id: "cond", text: "runes[i] != runes[len(runes)-1-i]", category: "condition", minWeight: "medium", weightScore: 0.8,
					hint: "比较第 i 个与倒数第 i 个字符，不对称返回 false", fillType: "text"},
				{id: "init", text: "i := 0", category: "init", minWeight: "beginner", weightScore: 0.4,
					hint: "循环起始下标（occurrence：for 循环的初始化）", fillType: "text"}, 
				{id: "incr", text: "i++", category: "iteration", minWeight: "easy", weightScore: 0.4,
					hint: "每次循环下标加一", fillType: "text"},
				{id: "arg", text: "\"hello\"", category: "argument", minWeight: "hard", weightScore: 0.5,
					hint: "第二个测试用例的反例字符串", fillType: "text", occurrence: 0},
			},
			hints: model.SpotterHints{
				Level1: "💡 首尾字符逐一比较",
				Level2: "📖 `runes[i]` 与 `runes[len(runes)-1-i]` 对称比较，不等则 false",
				Level3: "👀 `if runes[i] != runes[len(runes)-1-i] { return false }`",
				Level4: "📝 正确答案：`runes[i] != runes[len(runes)-1-i]`",
			},
			explanation: "双指针式循环：`runes[i]` 是首部字符，`runes[len(runes)-1-i]` 是对称尾部字符，任何一对不等即非回文。",
		},
		{
			id:          "go-factorial-bug-103",
			title:       "Goto 阶乘 Bug 修复",
			description: "修复阶乘计算的 off-by-one 与初始值 bug",
			exerciseType: model.ExerciseTypeBugfix,
			techStack:   "go",
			weight:      "hard",
			category:    "bugfix",
			fullCode:    "package main\n\nimport \"fmt\"\n\nfunc factorial(n int) int {\n\tresult := 1\n\tfor i := 1; i <= n; i++ {\n\t\tresult *= i\n\t}\n\treturn result\n}\n\nfunc main() {\n\tfmt.Println(factorial(5))\n\tfmt.Println(factorial(3))\n}",
			buggyCode:   "package main\n\nimport \"fmt\"\n\nfunc factorial(n int) int {\n\tresult := 1\n\tfor i := 0; i <= n; i++ {\n\t\tresult *= i\n\t}\n\treturn result\n}\n\nfunc main() {\n\tfmt.Println(factorial(5))\n\tfmt.Println(factorial(3))\n}",
			bugDesc:     "n=5 时输出 0，预期 120：循环从 i:=0 开始，第一次 result 乘以 0 后恒为 0",
			bugLoc:      "for i := 0; i <= n; i++",
			expected:    "120\n6",
			testCases:   []model.TestCase{{Output: "120"}, {Output: "6"}},
			points: []seedPoint{
				{id: "loop", text: "for i := 0; i <= n; i++", category: "loop", minWeight: "beginner", weightScore: 1,
					hint: "循环从哪个数开始？乘以 0 会发生什么？", fillType: "code", occurrence: 0},
			},
			hints: model.SpotterHints{
				Level1: "💡 循环起始值可能是问题所在",
				Level2: "📖 i 从 0 开始，第一次 result *= 0 直接把结果清零",
				Level3: "👀 改为 `for i := 1; i <= n; i++`",
				Level4: "📝 正确答案：`for i := 1; i <= n; i++`",
			},
			explanation: "Go 的 for 循环三条语句用分号分隔；`i := 0` 使第一次迭代 result 乘 0，结果恒为 0，应改为 `i := 1; i <= n; i++`。",
		},
		{
			id:          "go-fizzbuzz-104",
			title:       "Goz 思路 FizzBuzz",
			description: "实现 fizzbuzz：3 的倍数输出 Fizz，5 的倍数输出 Buzz，同倍数输出 FizzBuzz",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "go",
			weight:      "medium",
			category:    "control_flow",
			fullCode:    "package main\n\nimport \"fmt\"\n\nfunc fizzbuzz(n int) string {\n\tif n%15 == 0 {\n\t\treturn \"FizzBuzz\"\n\t}\n\tif n%3 == 0 {\n\t\treturn \"Fizz\"\n\t}\n\tif n%5 == 0 {\n\t\treturn \"Buzz\"\n\t}\n\treturn fmt.Sprint(n)\n}\n\nfunc main() {\n\tfmt.Println(fizzbuzz(15))\n\tfmt.Println(fizzbuzz(9))\n\tfmt.Println(fizzbuzz(10))\n\tfmt.Println(fizzbuzz(7))\n}",
			expected:    "FizzBuzz\nFizz\nBuzz\n7",
			points: []seedPoint{
				{id: "both", text: "n%15 == 0", category: "condition", minWeight: "beginner", weightScore: 0.8,
					hint: "同时是 3 和 5 的倍数，用最小公倍数 15 判断", fillType: "text"},
				{id: "fizz", text: "n%3 == 0", category: "condition", minWeight: "easy", weightScore: 0.5,
					hint: "3 的倍数", fillType: "text"},
				{id: "buzz", text: "n%5 == 0", category: "condition", minWeight: "easy", weightScore: 0.5,
					hint: "5 的倍数", fillType: "text"},
				{id: "fallback", text: "fmt.Sprint(n)", category: "expression", minWeight: "medium", weightScore: 0.5,
					hint: "都不是倍数时转成字符串返回", fillType: "text"},
			},
			hints: model.SpotterHints{
				Level1: "💡 三个分支都做取模判断",
				Level2: "📖 15、3、5 的倍数顺序判断，最后 `fmt.Sprint(n)`",
				Level3: "👀 `if n%15 == 0` / `if n%3 == 0` / `if n%5 == 0`",
				Level4: "📝 正确答案：`n%15==0` / `n%3==0` / `n%5==0` / `fmt.Sprint(n)`",
			},
			explanation: "先判 15 的倍数（同时满足 3 和 5），再分别判 3、5，最后返回字符串化的数字，注意 Go 里 int 不能直接与 string 拼接。",
		},
		{
			id:          "go-fib-105",
			title:       "Goz 斐波那契数列",
			description: "用递归实现 fib，返回第 n 个斐波那契数",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "go",
			weight:      "medium",
			category:    "recursion",
			fullCode:    "package main\n\nimport \"fmt\"\n\nfunc fib(n int) int {\n\tif n < 2 {\n\t\treturn n\n\t}\n\treturn fib(n-1) + fib(n-2)\n}\n\nfunc main() {\n\tfmt.Println(fib(10))\n}",
			expected:    "55",
			points: []seedPoint{
				{id: "base", text: "n < 2", category: "condition", minWeight: "beginner", weightScore: 0.6,
					hint: "递归终止条件：n 小于几时直接返回？", fillType: "text"},
				{id: "recur", text: "fib(n-1) + fib(n-2)", category: "expression", minWeight: "medium", weightScore: 0.9,
					hint: "斐波那契递推式：前两项之和", fillType: "text"},
				{id: "call", text: "10", category: "argument", minWeight: "easy", weightScore: 0.4,
					hint: "主程序求解第几个斐波那契数？", fillType: "text", occurrence: 0},
			},
			hints: model.SpotterHints{
				Level1: "💡 递归 = 终止条件 + 递推式",
				Level2: "📖 `n<2` 返回 n；否则 `fib(n-1)+fib(n-2)`",
				Level3: "👀 `if n < 2 { return n }` 否则 `return fib(n-1) + fib(n-2)`",
				Level4: "📝 正确答案：`n < 2` / `fib(n-1) + fib(n-2)` / `10`",
			},
			explanation: "Fib(0)=0、Fib(1)=1，其余为前两项之和；`fib(10)=55`。",
		},
		{
			id:          "go-concurrent-106",
			title:       "Gocon 并发求和",
			description: "用 goroutine 并发计算 1..100 的和，练习 sync.WaitGroup 与 Mutex",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "go",
			weight:      "hard",
			category:    "concurrency",
			fullCode:    "package main\n\nimport (\n\t\"fmt\"\n\t\"sync\"\n)\n\nfunc main() {\n\tvar wg sync.WaitGroup\n\tvar mu sync.Mutex\n\tsum := 0\n\tfor i := 1; i <= 100; i++ {\n\t\twg.Add(1)\n\t\tgo func(n int) {\n\t\t\tdefer wg.Done()\n\t\t\tmu.Lock()\n\t\t\tdefer mu.Unlock()\n\t\t\tsum += n\n\t\t}(i)\n\t}\n\twg.Wait()\n\tfmt.Println(sum)\n}",
			expected:    "5050",
			points: []seedPoint{
				{id: "add", text: "wg.Add(1)", category: "statement", minWeight: "beginner", weightScore: 0.5,
					hint: "每启动一个 goroutine 前计数器加一", fillType: "text"},
				{id: "done", text: "defer wg.Done()", category: "statement", minWeight: "easy", weightScore: 0.6,
					hint: "goroutine 结束时计数器减一", fillType: "text"},
				{id: "lock", text: "mu.Lock()", category: "statement", minWeight: "medium", weightScore: 0.5,
					hint: "并发修改 sum 前加互斥锁", fillType: "text"},
				{id: "goroutine", text: "go func", category: "keyword", minWeight: "easy", weightScore: 0.4,
					hint: "启动 goroutine 的关键字", fillType: "text"},
			},
			hints: model.SpotterHints{
				Level1: "💡 WaitGroup 管等待，Mutex 管互斥",
				Level2: "📖 `wg.Add(1)` 启动前增加，`defer wg.Done()` 退出前减少，写 sum 前 `mu.Lock()`",
				Level3: "👀 `go func(n int){ defer wg.Done(); mu.Lock(); ... }(i)`",
				Level4: "📝 正确答案：`wg.Add(1)` / `defer wg.Done()` / `mu.Lock()` / `go func`",
			},
			explanation: "`go` 关键字启动 goroutine；`WaitGroup` 等待所有 goroutine 完成；`Mutex` 保证并发写 sum 互斥。1..100 求和为 5050。",
		},
		{
			id:          "go-wordcount-107",
			title:       "Goword 词频统计",
			description: "用 map 统计一段文本中单词出现的次数",
			exerciseType: model.ExerciseTypeFill,
			techStack:   "go",
			weight:      "easy",
			category:    "map",
			fullCode:    "package main\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc main() {\n\ttext := \"go go python go go\"\n\twords := strings.Fields(text)\n\tcounts := map[string]int{}\n\tfor _, w := range words {\n\t\tcounts[w]++\n\t}\n\tfmt.Println(counts[\"go\"], counts[\"python\"], counts[\"rust\"])\n}",
			expected:    "4 1 0",
			points: []seedPoint{
				{id: "fields", text: "strings.Fields", category: "function", minWeight: "beginner", weightScore: 0.5,
					hint: "按空白切分字符串成切片的内置函数", fillType: "text"},
				{id: "incr", text: "counts[w]++", category: "statement", minWeight: "easy", weightScore: 0.6,
					hint: "单词出现次数加一（map 写入）", fillType: "text"},
				{id: "lookup", text: "counts[\"python\"]", category: "expression", minWeight: "medium", weightScore: 0.5,
					hint: "读取 python 出现的次数", fillType: "text"},
			},
			hints: model.SpotterHints{
				Level1: "💡 切分文本 + map 计数",
				Level2: "📖 `strings.Fields(text)` 切词，`counts[w]++` 计数",
				Level3: "👀 `words := strings.Fields(text)` 然后 `counts[w]++`",
				Level4: "📝 正确答案：`strings.Fields` / `counts[w]++` / `counts[\"python\"]`",
			},
			explanation: "`strings.Fields` 把字符串按空格切成切片；`map[string]int` 统计词频。",
		},
		{
			id:          "go-map-missing-108",
			title:       "Gomap map 键缺失处理 Bug 修复",
			description: "读取 map 中不存在的键时会返回零值并设置 ok=false，修复忽略 ok 的 bug",
			exerciseType: model.ExerciseTypeBugfix,
			techStack:   "go",
			weight:      "medium",
			category:    "bugfix",
			fullCode:    "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tscores := map[string]int{\"alice\": 90, \"bob\": 80}\n\tscore, ok := scores[\"carol\"]\n\tif !ok {\n\t\tfmt.Println(\"not found\")\n\t} else {\n\t\tfmt.Println(score)\n\t}\n}",
			buggyCode:   "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tscores := map[string]int{\"alice\": 90, \"bob\": 80}\n\tscore, _ := scores[\"carol\"]\n\tfmt.Println(score)\n}",
			bugDesc:     "carol 不在 map 中，忽略 ok 会直接输出零值 0，期望输出 not found",
			bugLoc:      "score, _ := scores[\"carol\"]",
			expected:    "not found",
			testCases:   []model.TestCase{{Output: "not found"}},
			points: []seedPoint{
				{id: "read", text: "score, _ := scores[\"carol\"]", category: "statement", minWeight: "beginner", weightScore: 1,
					hint: "map 读取返回两个值，第二个是键是否存在的布尔值，忽略它就会漏判", fillType: "code", occurrence: 0},
			},
			hints: model.SpotterHints{
				Level1: "💡 map 返回两个值：值和是否存在",
				Level2: "📖 用 `ok` 接收第二个返回值，`if !ok` 时输出 not found",
				Level3: "👀 `score, ok := scores[\"carol\"]; if !ok { fmt.Println(\"not found\") }`",
				Level4: "📝 正确答案：`score, ok := scores[\"carol\"]` 并判断 `if !ok`",
			},
			explanation: "Go 中读取 map 缺失键返回零值且 ok=false；必须检查 ok 才能真正区分「值存在但为 0」与「键不存在」。",
		},
	}

	out := make([]model.Exercise, 0, len(seedExercises)+50+50)
	for _, s := range seedExercises {
		ex, err := buildExercise(s)
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	// Kubernetes 基础题库（50 题）
	k8s, err := K8sBasics()
	if err != nil {
		return nil, err
	}
	out = append(out, k8s...)
	// Go 基础题库（50 题）
	gob, err := GoBasics()
	if err != nil {
		return nil, err
	}
	out = append(out, gob...)
	// Go 并发题库（50 题）
	gcc, err := GoConcurrency()
	if err != nil {
		return nil, err
	}
	out = append(out, gcc...)
	// Python 基础题库（50 题）
	pyb, err := PythonBasics()
	if err != nil {
		return nil, err
	}
	out = append(out, pyb...)
	// OpenShift 基础题库（50 题）
	ocb, err := OpenShiftBasics()
	if err != nil {
		return nil, err
	}
	out = append(out, ocb...)
	// OpenShift 对比 Kubernetes（30 题）
	ocd, err := OpenShiftDiff()
	if err != nil {
		return nil, err
	}
	out = append(out, ocd...)
	// Docker 基础题库（50 题）
	dkb, err := DockerBasics()
	if err != nil {
		return nil, err
	}
	out = append(out, dkb...)
	// Kubernetes 资源配置系列（30 题）
	res, err := K8sResourceConfig()
	if err != nil {
		return nil, err
	}
	out = append(out, res...)
	return out, nil
}
