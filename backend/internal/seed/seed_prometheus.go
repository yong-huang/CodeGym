package seed

import (
	"codegym/backend/internal/model"
)

// toPromSeed converts a Prometheus seed under the "prometheus"/"基础" stack.
func toPromSeed(k k8sBasicSeed) seedExercise {
	exType := model.ExerciseTypeCommand
	if k.kind == "config" {
		exType = model.ExerciseTypeConfig
	}
	return seedExercise{
		id: k.id, title: k.title, description: k.desc, exerciseType: exType,
		techStack: "prometheus", weight: k.weight, category: k.category, subcategory: "基础",
		fullCode: k.code, expected: k.expected, points: k.points,
		hints: goHint(k.hintL1, k.hintL2, k.hintL3, "正确答案："+k.code), explanation: k.explanation,
	}
}

// PrometheusBasics returns the 30-question Prometheus bank (promtool/API + PromQL + YAML).
func PrometheusBasics() ([]model.Exercise, error) {
	pt := func(id, text, category, weight, hint string) seedPoint {
		return seedPoint{id: id, text: text, category: category, minWeight: weight,
			weightScore: 1, hint: hint, fillType: "text"}
	}
	cmd := func(id, title, desc, category, code string, p seedPoint, l1, l2, l3, expl string) k8sBasicSeed {
		return k8sBasicSeed{id: id, title: title, desc: desc, kind: "command", category: category,
			weight: "easy", code: code, points: []seedPoint{p}, hintL1: l1, hintL2: l2, hintL3: l3,
			explanation: expl}
	}
	cfg := func(id, title, desc, category, code string, points []seedPoint, l1, l2, l3, _l4, expl string) k8sBasicSeed {
		return k8sBasicSeed{id: id, title: title, desc: desc, kind: "config", category: category,
			weight: "medium", code: code, points: points, hintL1: l1, hintL2: l2, hintL3: l3,
			explanation: expl}
	}

	seeds := []k8sBasicSeed{
		// ---- promtool / HTTP API ----
		cmd("prom-basics-001", "检查配置", "用 promtool 校验配置文件", "promtool",
			"promtool check config prometheus.yml",
			pt("verb", "check config", "command_verb", "beginner", "校验配置的子命令"),
			"校验配置的命令", "`promtool check config <文件>`", "`promtool check config prometheus.yml`",
			"`promtool check config` 校验配置语法与告警规则文件。"),

		cmd("prom-basics-002", "检查告警规则", "校验规则文件", "promtool",
			"promtool check rules rules.yml",
			pt("file", "rules.yml", "file", "easy", "规则文件名"),
			"校验规则的命令", "`promtool check rules <文件>`", "`promtool check rules rules.yml`",
			"`promtool check rules` 校验告警/记录规则语法。"),

		cmd("prom-basics-003", "promtool 查询", "用 promtool 执行 PromQL 查询", "promql",
			"promtool query 'up'",
			pt("query", "'up'", "promql", "easy", "PromQL 表达式"),
			"promtool 查询的命令", "`promtool query '<表达式>'`", "`promtool query 'up'`",
			"`promtool query` 对本地 Prometheus 执行即时查询，输出表达式结果。"),

		cmd("prom-basics-004", "即时查询 API", "用 HTTP API 查询即时值", "api",
			"curl 'http://localhost:9090/api/v1/query?query=up'",
			pt("q", "query=up", "api", "easy", "查询参数"),
			"即时查询的 API", "`/api/v1/query?query=<表达式>`", "`curl '.../api/v1/query?query=up'`",
			"`/api/v1/query` 返回某个时间点的即时查询结果。"),

		cmd("prom-basics-005", "健康检查", "查看 Prometheus 自身健康", "api",
			"curl http://localhost:9090/-/healthy",
			pt("path", "/-/healthy", "api", "easy", "健康检查端点"),
			"健康检查的端点", "`/-/healthy` 返回 200", "`curl http://localhost:9090/-/healthy`",
			"`/-/healthy` 是内置健康端点，返回 200 表示正常。"),

		cmd("prom-basics-006", "目标列表", "查看抓取目标", "api",
			"curl http://localhost:9090/api/v1/targets",
			pt("path", "/api/v1/targets", "api", "medium", "目标端点"),
			"查看抓取目标的 API", "`/api/v1/targets`", "`curl http://localhost:9090/api/v1/targets`",
			"`/api/v1/targets` 列出所有抓取目标及健康状态。"),

		cmd("prom-basics-007", "指标名列表", "查看所有指标名", "api",
			"curl http://localhost:9090/api/v1/label/__name__/values",
			pt("label", "__name__", "api", "medium", "指标名标签"),
			"查看指标名的 API", "`/api/v1/label/__name__/values`", "`curl .../api/v1/label/__name__/values`",
			"`__name__` 是指标名标签，values 返回全部指标名。"),

		cmd("prom-basics-008", "热重载配置", "不重启即可加载新配置", "api",
			"curl -X POST http://localhost:9090/-/reload",
			pt("method", "-X POST", "api", "medium", "HTTP 方法"),
			"热重载的命令", "`curl -X POST .../-/reload`", "`curl -X POST http://localhost:9090/-/reload`",
			"`/-/reload` 触发配置热重载，无需重启进程。"),

		cmd("prom-basics-009", "故障实例数", "统计 down 的目标", "promql",
			"promtool query 'count(up == 0)'",
			pt("expr", "up == 0", "promql", "medium", "过滤 down 的表达式"),
			"统计故障目标的表达式", "`count(up == 0)`", "`count(up == 0)`",
			"`up` 为 0 表示目标不可达，count 统计数量。"),

		cmd("prom-basics-010", "按指标名前缀匹配", "用正则匹配指标名", "promql",
			"promtool query '{__name__=~\"node_.*\"}'",
			pt("regex", "__name__=~\"node_.*\"", "promql", "hard", "正则匹配指标名"),
			"正则匹配指标名的表达式", "`{__name__=~\"node_.*\"}`", "`{__name__=~\"node_.*\"}`",
			"`__name__=~` 用正则匹配指标名前缀，选出所有 node_ 指标。"),

		// ---- PromQL 表达式 ----
		cmd("prom-basics-011", "增长率 rate", "计算计数器 5 分钟增长率", "promql",
			"rate(http_requests_total[5m])",
			pt("range", "[5m]", "promql", "easy", "时间范围"),
			"增长率的表达式", "`rate(指标[5m])` 每秒平均增量", "`rate(http_requests_total[5m])`",
			"`rate` 计算计数器在时间窗口内的每秒平均增长率，常用于 QPS。"),

		cmd("prom-basics-012", "按 job 求和", "聚合出每个 job 的请求量", "promql",
			"sum(rate(http_requests_total[5m])) by (job)",
			pt("by", "by (job)", "promql", "medium", "聚合维度"),
			"按 job 聚合的表达式", "`sum(...) by (job)`", "`sum(rate(http_requests_total[5m])) by (job)`",
			"`sum ... by (job)` 按 job 标签分组求和。"),

		cmd("prom-basics-013", "增量 increase", "统计 1 小时增量", "promql",
			"increase(node_network_receive_bytes_total[1h])",
			pt("fn", "increase", "promql", "medium", "增量函数"),
			"计算增量的函数", "`increase(指标[1h])` 时间窗内增量", "`increase(...[1h])`",
			"`increase` 返回计数器在窗口内增加的总量（近似）。"),

		cmd("prom-basics-014", "窗口平均 avg_over_time", "计算区间平均值", "promql",
			"avg_over_time(up[10m])",
			pt("fn", "avg_over_time", "promql", "hard", "区间聚合函数"),
			"区间平均的函数", "`avg_over_time(指标[窗口])`", "`avg_over_time(up[10m])`",
			"`avg_over_time` 对时间窗口内的所有样本取平均值。"),

		cmd("prom-basics-015", "直方图分位数", "计算请求延迟的 90 分位", "promql",
			"histogram_quantile(0.9, rate(http_request_duration_seconds_bucket[5m]))",
			pt("q", "0.9", "promql", "hard", "分位数"),
			"直方图分位数的表达式", "`histogram_quantile(0.9, rate(..._bucket[5m]))`", "`histogram_quantile(0.9, ...)`",
			"`histogram_quantile(0.9, ...)` 从直方图 bucket 计算 90 分位延迟。"),

		cmd("prom-basics-016", "请求错误率", "计算 5xx 占全部请求比例", "promql",
			"sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m]))",
			pt("filter", "status=~\"5..\"", "promql", "hard", "状态码正则过滤"),
			"错误率的表达式", "`{status=~\"5..\"}` 匹配 5xx 状态", "`sum(rate({status=~\"5..\"}[5m])) / sum(rate([5m]))`",
			"用正则 `status=~\"5..\"` 筛出 5xx，再除以总请求即错误率。"),

		cmd("prom-basics-017", "故障计数", "统计 up==0 的实例数", "promql",
			"count(up == 0)",
			pt("fn", "count", "promql", "medium", "计数函数"),
			"计数的函数", "`count(up == 0)` 统计故障数", "`count(up == 0)`",
			"`count` 统计满足条件的序列数量；`up == 0` 为故障目标。"),

		cmd("prom-basics-018", "按标签过滤", "只看指定实例的 up", "promql",
			"up{instance=\"host1:9100\"}",
			pt("sel", "{instance=\"host1:9100\"}", "promql", "easy", "标签选择器"),
			"按实例过滤的表达式", "`up{instance=\"host1:9100\"}`", "`up{instance=\"host1:9100\"}`",
			"`{标签=值}` 选择器按标签过滤序列。"),

		cmd("prom-basics-019", "布尔比较", "过滤内存使用大于 1G 的节点", "promql",
			"node_memory_used_bytes > 1e9",
			pt("cmp", "> 1e9", "promql", "easy", "比较表达式"),
			"比较过滤的表达式", "`指标 > 阈值`", "`node_memory_used_bytes > 1e9`",
			"PromQL 支持 `>、<、==` 比较，保留满足条件的序列。"),

		cmd("prom-basics-020", "求最大值", "取每个实例的最大请求率", "promql",
			"max(rate(http_requests_total[5m])) by (instance)",
			pt("fn", "max", "promql", "easy", "最大值函数"),
			"求最大的函数", "`max(...) by (instance)`", "`max(rate(http_requests_total[5m])) by (instance)`",
			"`max` 聚合取最大值，`by (instance)` 按实例分组。"),

		// ---- 配置 YAML / 参数 ----
		cmd("prom-basics-021", "保留时间", "设置数据保留 15 天", "storage",
			"prometheus --storage.tsdb.retention.time=15d",
			pt("flag", "--storage.tsdb.retention.time=15d", "flag", "medium", "保留时间参数"),
			"设置保留时间的参数", "`--storage.tsdb.retention.time=15d`", "`prometheus --storage.tsdb.retention.time=15d`",
			"用 `--storage.tsdb.retention.time` 控制 TSDB 数据保留时长。"),

		cfg("prom-basics-022", "global 抓取间隔", "配置全局抓取周期", "config",
			"global:\n  scrape_interval: 15s\n  evaluation_interval: 15s",
			[]seedPoint{
				pt("scrape", "15s", "interval", "easy", "抓取间隔"),
			},
			"`15s`", "global 抓取间隔", "`scrape_interval` 决定多久抓一次目标", "`scrape_interval: 15s`",
			"`global.scrape_interval` 是默认抓取周期，可被 job 覆盖。"),

		cfg("prom-basics-023", "抓取任务与目标", "定义一个抓取 job 及目标", "config",
			"scrape_configs:\n- job_name: node\n  static_configs:\n  - targets:\n    - localhost:9100",
			[]seedPoint{
				pt("job", "node", "job", "easy", "任务名"),
				pt("target", "localhost:9100", "target", "easy", "抓取目标"),
			},
			"`node` / `localhost:9100`", "scrape_configs", "`job_name` 分组抓取，`targets` 列目标", "`targets: [localhost:9100]`",
			"`scrape_configs` 定义抓取任务，`static_configs.targets` 列出地址。"),

		cfg("prom-basics-024", "metrics_path", "抓取自定义指标路径", "config",
			"scrape_configs:\n- job_name: app\n  metrics_path: /metrics\n  static_configs:\n  - targets:\n    - localhost:8080",
			[]seedPoint{
				pt("path", "/metrics", "path", "easy", "指标路径"),
			},
			"`/metrics`", "metrics_path", "默认抓 `/metrics`，可改自定义路径", "`metrics_path: /metrics`",
			"`metrics_path` 指定抓取指标端点路径，默认 `/metrics`。"),

		cfg("prom-basics-025", "规则文件", "引用告警/记录规则文件", "config",
			"rule_files:\n- \"rules/*.yml\"",
			[]seedPoint{
				pt("files", "\"rules/*.yml\"", "file", "easy", "规则文件路径"),
			},
			"`\"rules/*.yml\"`", "rule_files", "`rule_files` 加载告警与记录规则", "`rule_files: [\"rules/*.yml\"]`",
			"`rule_files` 列出规则文件，Prometheus 启动时加载。"),

		cfg("prom-basics-026", "告警通知", "配置 Alertmanager 地址", "config",
			"alerting:\n  alertmanagers:\n  - static_configs:\n    - targets:\n      - localhost:9093",
			[]seedPoint{
				pt("am", "localhost:9093", "alertmanager", "medium", "Alertmanager 地址"),
			},
			"`localhost:9093`", "alerting 配置", "`alerting.alertmanagers` 指定通知接收方", "`targets: [localhost:9093]`",
			"`alerting.alertmanagers` 告诉 Prometheus 把告警发给哪个 Alertmanager（默认 9093）。"),

		cfg("prom-basics-027", "监控自身", "抓取 Prometheus 自己的指标", "config",
			"scrape_configs:\n- job_name: prometheus\n  static_configs:\n  - targets:\n    - localhost:9090",
			[]seedPoint{
				pt("job", "prometheus", "job", "easy", "任务名"),
			},
			"`prometheus`", "自监控", "默认样例自带抓取 Prometheus 自身的 job", "`job_name: prometheus`",
			"自监控 job 抓取 Prometheus 自身的 `/metrics`（默认 9090）。"),

		cfg("prom-basics-028", "标签过滤抓取", "抓取时过滤标签", "config",
			"scrape_configs:\n- job_name: web\n  static_configs:\n  - targets:\n    - localhost:8080\n    - localhost:8081\n  relabel_configs:\n  - source_labels: [__address__]\n    regex: \".*:8080\"\n    action: keep",
			[]seedPoint{
				pt("action", "keep", "relabel", "hard", "保留匹配项的动作"),
			},
			"`keep`", "relabel 过滤", "`relabel_configs` 在抓取前改写/过滤标签", "`action: keep`",
			"`relabel_configs` 用 keep 只保留匹配 8080 的目标。"),

		cfg("prom-basics-029", "scrape_timeout", "设置抓取超时", "config",
			"global:\n  scrape_timeout: 10s\n  evaluation_interval: 15s",
			[]seedPoint{
				pt("timeout", "10s", "timeout", "easy", "抓取超时"),
			},
			"`10s`", "scrape_timeout", "`scrape_timeout` 限制单次抓取耗时", "`scrape_timeout: 10s`",
			"`global.scrape_timeout` 默认 10s，超时目标会标为 unhealthy。"),

		cfg("prom-basics-030", "外部标签", "抓取数据附加全局标签", "config",
			"global:\n  external_labels:\n    cluster: prod\n  scrape_interval: 15s",
			[]seedPoint{
				pt("cluster", "prod", "labels", "medium", "外部标签值"),
			},
			"`cluster: prod`", "external_labels", "`external_labels` 给所有序列附加标签", "`external_labels: {cluster: prod}`",
			"`global.external_labels` 为所有指标附加固定的全局标签。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, k := range seeds {
		ex, err := buildExercise(toPromSeed(k))
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
