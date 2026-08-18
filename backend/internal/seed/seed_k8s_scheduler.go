package seed

import (
	"codegym/backend/internal/model"
)

// toK8sScheduleSeed converts a scheduler seed under "kubernetes"/"调度".
func toK8sScheduleSeed(k k8sBasicSeed) seedExercise {
	exType := model.ExerciseTypeCommand
	if k.kind == "config" {
		exType = model.ExerciseTypeConfig
	}
	points := k.points
	if exType == model.ExerciseTypeCommand {
		points = addVerbPoint(points, k.code)
	}
	return seedExercise{
		id: k.id, title: k.title, description: k.desc, exerciseType: exType,
		techStack: "kubernetes", weight: k.weight, category: k.category, subcategory: "调度",
		fullCode: k.code, expected: k.expected, points: points,
		hints: goHint(k.hintL1, k.hintL2, k.hintL3, "正确答案："+k.code), explanation: k.explanation,
	}
}

// K8sScheduler returns the 30-question Kubernetes scheduler bank (调度).
func K8sScheduler() ([]model.Exercise, error) {
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
		// ---- kubectl 命令 ----
		cmd("k8s-sched-001", "查看调度器组件", "查看 kube-system 下的调度器 Pod", "scheduler",
			"kubectl get pods -n kube-system -l component=kube-scheduler",
			pt("flag", "-l component=kube-scheduler", "command_flag", "easy", "按组件标签筛选"),
			"查看调度器的命令", "`-l component=kube-scheduler` 选中调度器 Pod", "`kubectl get pods -n kube-system -l component=kube-scheduler`",
			"kube-scheduler 以 Pod 形式运行在 kube-system，用标签筛选查看。"),

		cmd("k8s-sched-002", "给节点打污点", "给 node-1 打 NoSchedule 污点", "taint",
			"kubectl taint nodes node-1 gpu=true:NoSchedule",
			pt("effect", "NoSchedule", "taint_effect", "easy", "污点效果"),
			"污点的 effect", "NoSchedule 阻止无容忍 Pod 调度", "`gpu=true:NoSchedule`",
			"污点 `key=value:effect`，effect 用 NoSchedule 阻止调度。"),

		cmd("k8s-sched-003", "节点不可调度", "将 node-1 标记为不可调度", "scheduling",
			"kubectl cordon node-1",
			pt("verb", "cordon", "command_verb", "easy", "标记不可调度"),
			"标记不可调度的命令", "`kubectl cordon <节点>`", "`kubectl cordon node-1`",
			"`cordon` 给节点打 unschedulable，新 Pod 不再调度到它。"),

		cmd("k8s-sched-004", "恢复节点调度", "取消 node-1 的不可调度状态", "scheduling",
			"kubectl uncordon node-1",
			pt("verb", "uncordon", "command_verb", "easy", "恢复调度"),
			"恢复调度的命令", "`kubectl uncordon <节点>`", "`kubectl uncordon node-1`",
			"`uncordon` 移除 unschedulable，节点重新接收调度。"),

		cmd("k8s-sched-005", "驱逐节点 Pod", "排空节点以便维护", "scheduling",
			"kubectl drain node-1 --ignore-daemonsets",
			pt("flag", "--ignore-daemonsets", "command_flag", "medium", "忽略 DaemonSet"),
			"驱逐 Pod 的选项", "`--ignore-daemonsets` 保留 DaemonSet", "`kubectl drain node-1 --ignore-daemonsets`",
			"`drain` 驱逐节点上所有可迁移 Pod，常配合 cordon 使用。"),

		cmd("k8s-sched-006", "查看 Pod 所在节点", "用 wide 输出查看 NODE 列", "scheduling",
			"kubectl get pods -o wide",
			pt("flag", "-o wide", "command_flag", "easy", "宽表输出"),
			"查看节点列的命令", "`-o wide` 显示 NODE/IP 列", "`kubectl get pods -o wide`",
			"`-o wide` 展示 Pod 被调度到的节点，排障调度结果常用。"),

		cmd("k8s-sched-007", "查看调度器日志", "查看 kube-scheduler 日志", "scheduler",
			"kubectl logs -n kube-system -l component=kube-scheduler",
			pt("flag", "-l component=kube-scheduler", "command_flag", "medium", "按组件标签选日志"),
			"查看调度器日志的命令", "`-l component=kube-scheduler` 选中调度器", "`kubectl logs -n kube-system -l component=kube-scheduler`",
			"调度器日志会记录绑定与调度失败原因。"),

		cmd("k8s-sched-008", "查看节点标签", "用 show-labels 列出节点标签", "node",
			"kubectl get nodes --show-labels",
			pt("flag", "--show-labels", "command_flag", "easy", "显示标签"),
			"查看节点标签的命令", "`--show-labels` 显示每节点标签", "`kubectl get nodes --show-labels`",
			"节点标签是 nodeSelector/亲和性的匹配依据。"),

		cmd("k8s-sched-009", "打节点标签", "给 node-1 打 disktype=ssd 标签", "node",
			"kubectl label nodes node-1 disktype=ssd",
			pt("label", "disktype=ssd", "labels", "easy", "标签键值"),
			"打节点标签的命令", "`kubectl label nodes <节点> <键>=<值>`", "`kubectl label nodes node-1 disktype=ssd`",
			"给节点打标签后，可用 nodeSelector 把 Pod 调度到这类节点。"),

		cmd("k8s-sched-010", "按节点过滤 Pod", "筛选调度到 node-1 的 Pod", "scheduling",
			"kubectl get pods --field-selector spec.nodeName=node-1",
			pt("flag", "--field-selector spec.nodeName=node-1", "command_flag", "hard", "按字段筛选"),
			"按节点筛选的命令", "`--field-selector spec.nodeName=<节点>`", "`kubectl get pods --field-selector spec.nodeName=node-1`",
			"用 field-selector 按 spec.nodeName 找出已调度到某节点的 Pod。"),

		cmd("k8s-sched-011", "查看 PriorityClass", "列出优先级类", "priority",
			"kubectl get priorityclasses",
			pt("res", "priorityclasses", "resource", "medium", "优先级类资源名"),
			"PriorityClass 资源名", "`kubectl get priorityclasses`", "`kubectl get priorityclasses`",
			"PriorityClass 决定调度顺序与抢占，值越大优先级越高。"),

		cmd("k8s-sched-012", "查看 Pod 调度详情", "describe 查看调度事件与节点", "scheduling",
			"kubectl describe pod web-1",
			pt("verb", "describe", "command_verb", "easy", "查看详情"),
			"查看调度详情的命令", "`kubectl describe pod <名>`", "`kubectl describe pod web-1`",
			"describe 里的 Events 会显示调度失败（如节点资源不足/污点）。"),

		cmd("k8s-sched-013", "查看节点资源用量", "查看节点资源占用", "node",
			"kubectl top node",
			pt("verb", "top", "command_verb", "medium", "查看资源用量"),
			"查看节点资源用量的命令", "`kubectl top node`", "`kubectl top node`",
			"`kubectl top node` 查看节点实际资源占用，判断是否资源不足导致调度失败。"),

		cmd("k8s-sched-014", "查看 Pending Pod", "筛选处于 Pending 的 Pod", "scheduling",
			"kubectl get pods -A --field-selector status.phase=Pending",
			pt("flag", "--field-selector status.phase=Pending", "command_flag", "hard", "按状态筛选"),
			"查看未调度 Pod 的命令", "`--field-selector status.phase=Pending`", "`kubectl get pods -A --field-selector status.phase=Pending`",
			"Pending 通常是调度器还没成功绑定（无可用节点/资源不足/污点）。"),

		// ---- YAML 配置 ----
		cfg("k8s-sched-015", "nodeSelector 定向调度", "只调度到 ssd 标签的节点", "affinity",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: ssd\nspec:\n  nodeSelector:\n    disktype: ssd\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("sel", "disktype: ssd", "selector", "easy", "节点标签键值"),
			},
			"`disktype: ssd`", "nodeSelector", "nodeSelector 匹配节点标签，只有带该标签的节点被选中", "`nodeSelector: {disktype: ssd}`",
			"`nodeSelector` 按节点标签定向调度，是最简单的调度约束。"),

		cfg("k8s-sched-016", "固定节点 nodeName", "直接把 Pod 绑到 node-1", "scheduling",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: pinned\nspec:\n  nodeName: node-1\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("node", "node-1", "node", "easy", "目标节点名"),
			},
			"`node-1`", "nodeName 直绑", "spec.nodeName 绕过调度器直接绑定节点", "`nodeName: node-1`",
			"`spec.nodeName` 直接指定节点，调度器跳过选择过程。"),

		cfg("k8s-sched-017", "required 节点亲和性", "硬性要求调度到 ssd 节点", "affinity",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: aff\nspec:\n  affinity:\n    nodeAffinity:\n      requiredDuringSchedulingIgnoredDuringExecution:\n        nodeSelectorTerms:\n        - matchExpressions:\n          - key: disktype\n            operator: In\n            values:\n            - ssd\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("key", "disktype", "labels", "medium", "匹配的节点标签键"),
				pt("op", "In", "operator", "hard", "匹配操作符"),
			},
			"`disktype` / `In`", "required 亲和性", "required 是硬约束，没有满足节点就不调度", "`operator: In`",
			"nodeAffinity 的 required 子句是硬性要求，比 nodeSelector 支持 In/NotIn 等操作符。"),

		cfg("k8s-sched-018", "preferred 节点亲和性", "优先调度到 ssd 节点但不强制", "affinity",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: pref\nspec:\n  affinity:\n    nodeAffinity:\n      preferredDuringSchedulingIgnoredDuringExecution:\n      - weight: 100\n        preference:\n          matchExpressions:\n          - key: disktype\n            operator: In\n            values:\n            - ssd\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("weight", "100", "weight", "hard", "偏好权重"),
			},
			"`100`", "preferred 亲和性", "preferred 是软偏好，权重影响但不强制", "`weight: 100`",
			"preferred 子句按 weight 加分，节点不满足也能调度，只是优先级低。"),

		cfg("k8s-sched-019", "Pod 亲和性", "与 web 标签的 Pod 同节点调度", "affinity",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: co\nspec:\n  affinity:\n    podAffinity:\n      requiredDuringSchedulingIgnoredDuringExecution:\n      - labelSelector:\n          matchExpressions:\n          - key: app\n            operator: In\n            values:\n            - web\n        topologyKey: kubernetes.io/hostname\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("topo", "kubernetes.io/hostname", "topology", "hard", "拓扑键"),
			},
			"`kubernetes.io/hostname`", "podAffinity", "与满足标签的 Pod 在相同拓扑域内调度", "`topologyKey: kubernetes.io/hostname`",
			"podAffinity 让新 Pod 与已有 Pod 靠近（同节点/同区域）。"),

		cfg("k8s-sched-020", "Pod 反亲和性", "避免与 web Pod 同节点", "affinity",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: anti\nspec:\n  affinity:\n    podAntiAffinity:\n      requiredDuringSchedulingIgnoredDuringExecution:\n      - labelSelector:\n          matchExpressions:\n          - key: app\n            operator: In\n            values:\n            - web\n        topologyKey: kubernetes.io/hostname\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("anti", "podAntiAffinity", "affinity", "hard", "反亲和字段"),
			},
			"`podAntiAffinity`", "podAntiAffinity", "避免与特定 Pod 在同一拓扑域，常用于分散副本", "`topologyKey: kubernetes.io/hostname`",
			"podAntiAffinity 让副本分散到不同节点/区域，提升可用性。"),

		cfg("k8s-sched-021", "容忍污点", "容忍 NoSchedule 污点以便调度", "taint",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: util\nspec:\n  tolerations:\n  - key: gpu\n    operator: Exists\n    effect: NoSchedule\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("effect", "NoSchedule", "taint_effect", "medium", "污点效果"),
				pt("op", "Exists", "operator", "hard", "容忍操作符"),
			},
			"`NoSchedule` / `Exists`", "tolerations", "容忍匹配污点的 key/effect 才能调度", "`effect: NoSchedule`",
			"Pod 声明 tolerations 匹配节点污点后，才能被调度到该节点。"),

		cfg("k8s-sched-022", "容忍 key 匹配", "只容忍指定 key 的污点", "taint",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: gpu\nspec:\n  tolerations:\n  - key: gpu\n    operator: Equal\n    value: \"true\"\n    effect: NoSchedule\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("key", "gpu", "taint_key", "medium", "污点键"),
				pt("val", "true", "taint_value", "hard", "污点值"),
			},
			"`gpu` / `\"true\"`", "tolerations 精确匹配", "Equal 操作符要求 key+value 都匹配", "`operator: Equal`",
			"Equal 容忍要求污点 key 与 value 都匹配，比 Exists 更精确。"),

		cfg("k8s-sched-023", "PriorityClass 定义", "定义高优先级类", "priority",
			"apiVersion: scheduling.k8s.io/v1\nkind: PriorityClass\nmetadata:\n  name: high-priority\nvalue: 1000000\nglobalDefault: false",
			[]seedPoint{
				pt("api", "scheduling.k8s.io/v1", "api_version", "easy", "PriorityClass 的 API 组"),
				pt("val", "1000000", "priority", "hard", "优先级数值"),
			},
			"`scheduling.k8s.io/v1` / `1000000`", "PriorityClass", "值越大优先级越高，可抢占低优先级 Pod", "`value: 1000000`",
			"PriorityClass 定义优先级，调度器先调度高优先级，必要时抢占。"),

		cfg("k8s-sched-024", "Pod 引用优先级", "指定 pod 使用 high-priority 类", "priority",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: hp\nspec:\n  priorityClassName: high-priority\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("pri", "high-priority", "priority", "medium", "优先级类名"),
			},
			"`high-priority`", "priorityClassName", "spec.priorityClassName 引用 PriorityClass", "`priorityClassName: high-priority`",
			"Pod 通过 priorityClassName 指定优先级，参与调度排队与抢占。"),

		cfg("k8s-sched-025", "拓扑分布约束", "副本按可用区分散", "topology",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: web\n  labels:\n    app: web\nspec:\n  topologySpreadConstraints:\n  - maxSkew: 1\n    topologyKey: topology.kubernetes.io/zone\n    whenUnsatisfiable: ScheduleAnyway\n    labelSelector:\n      matchLabels:\n        app: web\n  containers:\n  - name: web\n    image: nginx",
			[]seedPoint{
				pt("skew", "1", "topology", "hard", "最大不均衡数"),
				pt("key", "topology.kubernetes.io/zone", "topology", "medium", "拓扑键"),
			},
			"`1` / `topology.kubernetes.io/zone`", "TopologySpread", "控制副本在各拓扑域的分布，避免集中", "`maxSkew: 1`",
			"TopologySpreadConstraints 让副本按区域/节点均匀分布。"),

		cfg("k8s-sched-026", "资源请求（调度依据）", "声明 CPU 请求供调度", "resources",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: r\nspec:\n  containers:\n  - name: c\n    image: nginx\n    resources:\n      requests:\n        cpu: \"500m\"\n        memory: \"256Mi\"",
			[]seedPoint{
				pt("cpu", "500m", "resource", "easy", "CPU 请求"),
			},
			"`500m`", "resources.requests", "调度器按 requests 评估节点剩余容量", "`requests.cpu: \"500m\"`",
			"调度器只依据 requests（而非 limits）判断节点是否放得下。"),

		cfg("k8s-sched-027", "自定义调度器", "用自定义 schedulerName 调度", "scheduler",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: custom\nspec:\n  schedulerName: my-scheduler\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("name", "my-scheduler", "scheduler", "medium", "自定义调度器名"),
			},
			"`my-scheduler`", "schedulerName", "默认调度器是 default-scheduler，可指定其它调度器", "`schedulerName: my-scheduler`",
			"spec.schedulerName 指定自定义调度器，实现个性化的调度策略。"),

		cfg("k8s-sched-028", "DaemonSet 逐节点调度", "DaemonSet 在每个节点各跑一个", "daemonset",
			"apiVersion: apps/v1\nkind: DaemonSet\nmetadata:\n  name: fluentd\nspec:\n  selector:\n    matchLabels:\n      app: fluentd\n  template:\n    metadata:\n      labels:\n        app: fluentd\n    spec:\n      containers:\n      - name: fluentd\n        image: fluent/fluentd",
			[]seedPoint{
				pt("kind", "DaemonSet", "kind", "easy", "资源类型"),
			},
			"`DaemonSet`", "DaemonSet 调度", "DaemonSet 在每个节点各放一个 Pod，不经过常规调度选择", "`kind: DaemonSet`",
			"DaemonSet 由控制器保证每节点一个，调度器不做放置决策。"),

		cfg("k8s-sched-029", "PodDisruptionBudget", "保证可用副本数", "pdb",
			"apiVersion: policy/v1\nkind: PodDisruptionBudget\nmetadata:\n  name: web-pdb\nspec:\n  minAvailable: 2\n  selector:\n    matchLabels:\n      app: web",
			[]seedPoint{
				pt("min", "2", "pdb", "medium", "最少可用副本数"),
			},
			"`2`", "PodDisruptionBudget", "PDB 限制自愿驱逐（节点维护）时可同时不可用数量", "`minAvailable: 2`",
			"PDB 保证自愿中断（如 drain）时仍有足够副本在线。"),

		cfg("k8s-sched-030", "污点 + 容忍组合", "节点污点与 Pod 容忍配合", "taint",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: dedicated\nspec:\n  nodeSelector:\n    dedicated: infra\n  tolerations:\n  - key: dedicated\n    operator: Exists\n    effect: NoSchedule\n  containers:\n  - name: c\n    image: nginx",
			[]seedPoint{
				pt("tol", "dedicated", "taint_key", "hard", "容忍的污点键"),
			},
			"`dedicated`", "污点 + nodeSelector", "节点打污点拒绝普通 Pod，专用 Pod 用容忍 + 标签定向", "`key: dedicated`",
			"污点阻止默认 Pod，专用 Pod 用 tolerations 匹配并通过 nodeSelector 定向。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, k := range seeds {
		ex, err := buildExercise(toK8sScheduleSeed(k))
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
