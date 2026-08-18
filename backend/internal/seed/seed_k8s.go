package seed

import (
	"codegym/backend/internal/model"
)

// k8sBasicSeed declares one Kubernetes basic question. kind is "command"
// (kubectl 命令补全) or "config" (YAML 字段补全).
type k8sBasicSeed struct {
	id          string
	title       string
	desc        string
	kind        string
	category    string
	weight      string
	code        string
	expected    string
	points      []seedPoint
	hintL1      string
	hintL2      string
	hintL3      string
	explanation string
}

// spotter builds the 4-level hint ladder, revealing the answer at L4.
func spotter(answer, l1, l2, l3 string) model.SpotterHints {
	return model.SpotterHints{
		Level1: "💡 " + l1,
		Level2: "📖 " + l2,
		Level3: "👀 " + l3,
		Level4: "📝 正确答案：" + answer,
	}
}

func toSeedExercise(k k8sBasicSeed) seedExercise {
	exType := model.ExerciseTypeCommand
	if k.kind == "config" {
		exType = model.ExerciseTypeConfig
	}
	points := k.points
	if exType == model.ExerciseTypeCommand {
		points = addVerbPoint(points, k.code)
	}
	return seedExercise{
		id:           k.id,
		title:        k.title,
		description:  k.desc,
		exerciseType: exType,
		techStack:    "kubernetes",
		weight:       k.weight,
		category:     k.category,
		fullCode:     k.code,
		expected:     k.expected,
		points:       points,
		hints:        spotter("", k.hintL1, k.hintL2, k.hintL3),
		explanation:  k.explanation,
	}
}

// K8sBasics returns the 50-question Kubernetes basic bank (基础题).
func K8sBasics() ([]model.Exercise, error) {
	// point shortcut: single text fill with acceptable answers
	pt := func(id, text, category, weight, hint string, fillType string, acc []string) seedPoint {
		return seedPoint{id: id, text: text, category: category, minWeight: weight, weightScore: 1,
			hint: hint, fillType: fillType, acceptableAnswers: acc}
	}

	cmd := func(id, title, desc, category, code string, p seedPoint, l1, l2, l3, expl string) k8sBasicSeed {
		return k8sBasicSeed{id: id, title: title, desc: desc, kind: "command", category: category,
			weight: "easy", code: code, points: []seedPoint{p}, hintL1: l1, hintL2: l2, hintL3: l3,
			explanation: expl}
	}
	cfg := func(id, title, desc, category, weight, code string, points []seedPoint, l1, l2, l3, expl string) k8sBasicSeed {
		return k8sBasicSeed{id: id, title: title, desc: desc, kind: "config", category: category,
			weight: weight, code: code, points: points, hintL1: l1, hintL2: l2, hintL3: l3,
			explanation: expl}
	}

	seeds := []k8sBasicSeed{
		cmd("k8s-basics-001", "查看所有命名空间的 Pod", "列出所有命名空间下的 Pod", "command_get",
			"kubectl get pods -A",
			pt("flag", "-A", "command_flag", "beginner", "短参数 -A 或长参数 --all-namespaces", "select", []string{"--all-namespaces"}),
			"所有命名空间的关键词是什么？", "`--all-namespaces` 或短参数 `-A`", "`kubectl get pods -A`",
			"`-A` 是 `--all-namespaces` 的短参数，列出全部命名空间的 Pod。"),
		cmd("k8s-basics-002", "查看指定命名空间的 Pod", "列出 kube-system 命名空间的 Pod", "command_get",
			"kubectl get pods -n kube-system",
			pt("flag", "-n kube-system", "command_flag", "beginner", "用 -n 或 --namespace 指定命名空间", "select", []string{"--namespace kube-system"}),
			"如何指定命名空间？", "`-n kube-system` 或 `--namespace kube-system`", "`kubectl get pods -n kube-system`",
			"`-n` 指定命名空间，不指定时默认使用当前 context 的命名空间。"),
		cmd("k8s-basics-003", "查看 Pod 详细信息", "查看名为 nginx 的 Pod 的详细信息与事件", "command_describe",
			"kubectl describe pod nginx",
			pt("verb", "describe", "command_verb", "beginner", "查看资源详细状态和事件的命令动词", "text", nil),
			"查看详细信息的命令是？", "`kubectl describe pod <名称>`", "`kubectl describe pod nginx`",
			"`describe` 展示资源的详细状态、事件与最近事件，排障常用。"),
		cmd("k8s-basics-004", "查看集群节点", "列出集群中的所有节点", "command_get",
			"kubectl get nodes",
			pt("resource", "nodes", "resource", "beginner", "节点的资源名", "text", nil),
			"节点的资源名是什么？", "`kubectl get nodes`", "`kubectl get nodes`",
			"节点资源名为 `nodes`，可用 `-o wide` 查看更多信息。"),
		cmd("k8s-basics-005", "应用 YAML 清单", "用 deployment.yaml 创建或更新资源", "command_apply",
			"kubectl apply -f deployment.yaml",
			pt("flag", "-f deployment.yaml", "command_flag", "beginner", "指定清单文件的选项", "text", nil),
			"用什么选项指定清单文件？", "`-f <文件名>`", "`kubectl apply -f deployment.yaml`",
			"`apply -f` 声明式应用清单，可创建或更新资源。"),
		cmd("k8s-basics-006", "删除 Pod", "删除名为 nginx 的 Pod", "command_delete",
			"kubectl delete pod nginx",
			pt("verb", "delete", "command_verb", "beginner", "删除资源的命令动词", "text", nil),
			"删除资源的动词？", "`kubectl delete pod <名称>`", "`kubectl delete pod nginx`",
			"`delete` 删除资源，可用 `--force --grace-period=0` 强制删除。"),
		cmd("k8s-basics-007", "创建命名空间", "创建名为 dev 的命名空间", "command_create",
			"kubectl create namespace dev",
			pt("verb", "create", "command_verb", "beginner", "创建资源的命令动词", "text", nil),
			"创建资源的动词？", "`kubectl create namespace dev`", "`kubectl create namespace dev`",
			"`create` 命令式创建资源；`apply` 则更适用于声明式管理。"),
		cmd("k8s-basics-008", "运行一次性 Pod", "以 nginx 镜像运行一个一次性 Pod（不生成控制器）", "command_run",
			"kubectl run nginx --image=nginx:1.25 --restart=Never",
			pt("restart", "--restart=Never", "command_flag", "medium", "让 kubectl run 创建裸 Pod 而非 Deployment 的选项", "text", nil),
			"如何让 run 创建裸 Pod？", "`--restart=Never` 创建裸 Pod", "`kubectl run nginx --image=nginx:1.25 --restart=Never`",
			"默认 `kubectl run` 创建 Deployment；`--restart=Never` 创建一次性裸 Pod。"),
		cmd("k8s-basics-009", "暴露 Deployment 服务", "将 nginx Deployment 暴露为 ClusterIP 服务", "command_expose",
			"kubectl expose deployment nginx --port=80 --type=ClusterIP",
			pt("type", "--type=ClusterIP", "command_flag", "easy", "服务类型的选项", "text", []string{"--type=NodePort"}),
			"服务类型参数是什么？", "`--type=ClusterIP` 或 `--type=NodePort`", "`kubectl expose deployment nginx --port=80 --type=ClusterIP`",
			"`expose` 为资源创建 Service；`--type` 决定 Service 暴露方式。"),
		cmd("k8s-basics-010", "查看 Deployment", "列出所有 Deployment", "command_get",
			"kubectl get deployments",
			pt("resource", "deployments", "resource", "beginner", "Deployment 的资源名", "text", []string{"deploy"}),
			"Deployment 的资源名（复数）？", "`deployments` 或简称 `deploy`", "`kubectl get deployments`",
			"资源名可用全称或缩写（deploy/svc/pods）。"),
		cmd("k8s-basics-011", "查看 Service", "列出所有 Service", "command_get",
			"kubectl get svc",
			pt("resource", "svc", "resource", "beginner", "Service 的缩写资源名", "text", []string{"services"}),
			"Service 的缩写资源名？", "`svc` 或 `services`", "`kubectl get svc`",
			"Service 缩写为 `svc`，全称为 `services`。"),
		cmd("k8s-basics-014", "查看集群事件", "查看命名空间内的最新事件", "command_events",
			"kubectl get events",
			pt("resource", "events", "resource", "easy", "事件资源名", "text", nil),
			"查看事件的资源名？", "`kubectl get events`", "`kubectl get events`",
			"`events` 记录调度、拉取镜像、重启等集群事件，排障重要。"),
		cmd("k8s-basics-015", "查看 Pod 日志", "查看 nginx Pod 的标准输出日志", "command_logs",
			"kubectl logs nginx",
			pt("verb", "logs", "command_verb", "beginner", "查看日志的命令动词", "text", nil),
			"查看日志的动词？", "`kubectl logs <pod>`", "`kubectl logs nginx`",
			"`logs` 输出容器 stdout/stderr 日志，`-f` 实时跟随。"),
		cmd("k8s-basics-016", "实时跟随日志", "实时输出 nginx Pod 的日志流", "command_logs",
			"kubectl logs -f nginx",
			pt("flag", "-f", "command_flag", "easy", "实时跟随日志的选项", "text", []string{"--follow"}),
			"跟随日志输出的选项？", "`-f` 或 `--follow`", "`kubectl logs -f nginx`",
			"`-f` 实时流式输出日志，类似 `tail -f`。"),
		cmd("k8s-basics-017", "进入容器执行命令", "以交互式终端进入 nginx 容器的 shell", "command_exec",
			"kubectl exec -it nginx -- /bin/sh",
			pt("args", "-it", "command_flag", "easy", "交互式 + 分配 TTY 的选项", "text", []string{"--stdin --tty"}),
			"交互式终端的选项组合？", "`-it`（stdin + tty）", "`kubectl exec -it nginx -- /bin/sh`",
			"`-i` 保持标准输入，`-t` 分配 TTY，`--` 后接要执行的命令。"),
		cmd("k8s-basics-018", "端口转发到本地", "将 nginx Service 的 80 端口转发到本地 8080", "command_portforward",
			"kubectl port-forward svc/nginx 8080:80",
			pt("ports", "8080:80", "command_flag", "easy", "本地端口:容器端口的格式", "text", nil),
			"端口映射的写法？", "`本地端口:目标端口` → `8080:80`", "`kubectl port-forward svc/nginx 8080:80`",
			"`port-forward` 把集群内端口转发到本地，便于调试访问。"),
		cmd("k8s-basics-019", "副本扩容", "将 nginx Deployment 扩到 3 个副本", "command_scale",
			"kubectl scale deployment nginx --replicas=3",
			pt("flag", "--replicas=3", "command_flag", "easy", "设置副本数的选项", "text", nil),
			"设置副本数的选项？", "`--replicas=3`", "`kubectl scale deployment nginx --replicas=3`",
			"`scale` 调整副本数；HPA 会自动管理副本数。"),
		cmd("k8s-basics-020", "更新镜像版本", "将 nginx Deployment 的镜像更新到 nginx:1.26", "command_set",
			"kubectl set image deployment/nginx nginx=nginx:1.26",
			pt("verb", "set image", "command_verb", "medium", "更新镜像的命令", "text", nil),
			"更新镜像的命令？", "`kubectl set image deployment/<名> <容器>=<镜像>`", "`kubectl set image deployment/nginx nginx=nginx:1.26`",
			"`set image` 滚动更新容器镜像；可用 `kubectl rollout status` 查看进度。"),
		cmd("k8s-basics-021", "查看滚动更新状态", "查看 nginx Deployment 的滚动更新状态", "command_rollout",
			"kubectl rollout status deployment/nginx",
			pt("verb", "rollout status", "command_verb", "easy", "查看更新状态的命令", "text", nil),
			"查看滚动状态的命令？", "`kubectl rollout status deployment/nginx`", "`kubectl rollout status deployment/nginx`",
			"`rollout status` 阻塞直到部署完成或超时。"),
		cmd("k8s-basics-022", "回滚更新", "将 nginx Deployment 回滚到上一个版本", "command_rollout",
			"kubectl rollout undo deployment/nginx",
			pt("verb", "undo", "command_verb", "easy", "回滚到上一版本的子命令", "text", nil),
			"回滚的子命令？", "`rollout undo`", "`kubectl rollout undo deployment/nginx`",
			"`rollout undo` 回滚到上一版本，`--to-revision` 指定目标版本。"),
		cmd("k8s-basics-025", "给资源加注解", "给 nginx Pod 添加描述注解", "command_annotate",
			"kubectl annotate pod nginx description=web-server",
			pt("verb", "annotate", "command_verb", "medium", "管理注解的命令动词", "text", nil),
			"管理注解的动词？", "`kubectl annotate pod <名> <key>=<value>`", "`kubectl annotate pod nginx description=web-server`",
			"`annotate` 添加非标识性元数据；注解不用于选择器匹配。"),
		cmd("k8s-basics-026", "以 YAML 输出", "以 YAML 格式输出 nginx Pod 的清单", "command_output",
			"kubectl get pod nginx -o yaml",
			pt("flag", "-o yaml", "command_flag", "easy", "YAML 输出格式选项", "text", []string{"--output yaml"}),
			"指定输出格式的选项？", "`-o yaml` 或 `--output yaml`", "`kubectl get pod nginx -o yaml`",
			"`-o yaml` 输出资源清单，便于审阅或另存复用。"),
		cmd("k8s-basics-027", "以 JSON 输出", "以 JSON 格式输出 nginx Pod 的清单", "command_output",
			"kubectl get pod nginx -o json",
			pt("flag", "-o json", "command_flag", "easy", "JSON 输出格式选项", "text", []string{"--output json"}),
			"JSON 输出格式选项？", "`-o json`", "`kubectl get pod nginx -o json`",
			"`-o json` 输出结构化 JSON，便于脚本处理。"),
		cmd("k8s-basics-028", "宽表输出", "以宽格式查看 Pod（含节点 IP）", "command_output",
			"kubectl get pods -o wide",
			pt("flag", "-o wide", "command_flag", "easy", "宽表输出选项", "text", nil),
			"显示更多列的选项？", "`-o wide`", "`kubectl get pods -o wide`",
			"`-o wide` 额外显示节点、IP 等列。"),
		cmd("k8s-basics-029", "按标签筛选", "筛选出 app=nginx 标签的 Pod", "command_select",
			"kubectl get pods -l app=nginx",
			pt("flag", "-l app=nginx", "command_flag", "easy", "标签选择器选项", "text", []string{"--selector app=nginx"}),
			"标签筛选的选项？", "`-l app=nginx` 或 `--selector`", "`kubectl get pods -l app=nginx`",
			"`-l` 按标签键值筛选，支持 `,`、`!=`、`in` 等表达式。"),
		cmd("k8s-basics-031", "客户端预演", "在不真正应用的情况下验证 pod.yaml 语法", "command_dryrun",
			"kubectl apply -f pod.yaml --dry-run=client",
			pt("flag", "--dry-run=client", "command_flag", "easy", "本地校验不落地的选项", "text", []string{"--dry-run=server"}),
			"预演选项？", "`--dry-run=client`（本地）或 `=server`（服务端）", "`kubectl apply -f pod.yaml --dry-run=client`",
			"`--dry-run` 校验清单而不真正变更集群。"),
		cmd("k8s-basics-036", "驱逐节点 Pod", "清空 node-1 上的 Pod 以便维护", "command_drain",
			"kubectl drain node-1 --ignore-daemonsets",
			pt("flag", "--ignore-daemonsets", "command_flag", "medium", "忽略 DaemonSet 的选项", "text", nil),
			"忽略 DaemonSet 的选项？", "`--ignore-daemonsets`", "`kubectl drain node-1 --ignore-daemonsets`",
			"`drain` 驱逐节点上 Pod（通常配合 cordon 使用）。"),
		cmd("k8s-basics-039", "配置水平自动伸缩", "让 nginx Deployment 在 2-5 副本间自动伸缩", "command_autoscale",
			"kubectl autoscale deployment nginx --min=2 --max=5",
			pt("flag", "--max=5", "command_flag", "medium", "设置最大副本数的选项", "text", []string{"--cpu-percent=80"}),
			"设置最大副本数的选项？", "`--min=2 --max=5`", "`kubectl autoscale deployment nginx --min=2 --max=5`",
			"`autoscale` 创建 HPA，按 CPU 使用率自动调整副本数。"),
		cmd("k8s-basics-040", "查看当前上下文", "查看当前使用的 kubeconfig 上下文", "command_config",
			"kubectl config current-context",
			pt("verb", "current-context", "command_verb", "easy", "查看当前上下文的子命令", "text", nil),
			"查看当前上下文的命令？", "`kubectl config current-context`", "`kubectl config current-context`",
			"上下文 = 集群 + 用户 + 命名空间的组合。"),
		cmd("k8s-basics-041", "切换上下文", "切换到 minikube 上下文", "command_config",
			"kubectl config use-context minikube",
			pt("verb", "use-context", "command_verb", "easy", "切换上下文的子命令", "text", nil),
			"切换上下文的命令？", "`kubectl config use-context <名称>`", "`kubectl config use-context minikube`",
			"`use-context` 切换当前生效的集群上下文。"),
		cfg("k8s-basics-051", "Pod 资源限制", "为容器设置 CPU 与内存限制", "yaml_limits", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: busybox\nspec:\n  containers:\n  - name: busybox\n    image: busybox\n    resources:\n      limits:\n        cpu: \"500m\"\n        memory: \"128Mi\"",
			[]seedPoint{
				pt("cpu", "\"500m\"", "resource", "easy", "CPU 限制（500m = 0.5 核）", "text", nil),
				pt("mem", "\"128Mi\"", "resource", "easy", "内存限制（Mi 为二进制兆）", "text", nil),
			},
			"limits 需要哪两项？", "`cpu` 与 `memory`，500m=0.5 核，128Mi=128 MiB", "`cpu: \"500m\"` / `memory: \"128Mi\"`",
			"resources.limits 设置容器可用的硬上限；requests 是调度保证。"),
		cfg("k8s-basics-052", "Pod 环境变量", "通过 env 注入环境变量", "yaml_env", "easy",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: web\nspec:\n  containers:\n  - name: web\n    image: nginx\n    env:\n    - name: MODE\n      value: production",
			[]seedPoint{
				pt("name", "MODE", "config", "beginner", "环境变量名", "text", nil),
				pt("value", "production", "config", "beginner", "环境变量值", "text", nil),
			},
			"env 每项的两个字段？", "`name` 与 `value`", "`- name: MODE` / `value: production`",
			"容器通过 env 注入键值配置；值也可来自 configMapKeyRef/secretKeyRef。"),
		cfg("k8s-basics-053", "Deployment 副本数", "创建 3 副本的 nginx Deployment", "yaml_deployment", "easy",
			"apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx-deploy\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.25",
			[]seedPoint{
				pt("api", "apps/v1", "api_version", "easy", "Deployment 的 API 版本", "text", nil),
				pt("replicas", "3", "replicas", "easy", "期望副本数", "text", nil),
			},
			"Deployment 的 apiVersion 与副本数？", "`apps/v1`，`replicas: 3`", "`apiVersion: apps/v1` / `replicas: 3`",
			"Deployment 属于 apps 组 v1；replicas 声明期望副本数，label 选择器需匹配模板标签。"),
		cfg("k8s-basics-054", "Deployment 更新策略", "配置滚动更新最大不可用为 25%", "yaml_deployment", "hard",
			"apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: app\nspec:\n  replicas: 2\n  strategy:\n    type: RollingUpdate\n    rollingUpdate:\n      maxUnavailable: 25%\n  selector:\n    matchLabels:\n      app: app\n  template:\n    metadata:\n      labels:\n        app: app\n    spec:\n      containers:\n      - name: app\n        image: nginx",
			[]seedPoint{
				pt("maxun", "maxUnavailable: 25%", "config", "medium", "滚动更新期间最多不可用副本比例", "text", nil),
			},
			"控制滚动更新速率的字段？", "`rollingUpdate.maxUnavailable`（或 maxSurge）", "`maxUnavailable: 25%`",
			"`maxSurge` 与 `maxUnavailable` 共同控制滚动更新的节奏。"),
		cfg("k8s-basics-055", "Service 类型与端口", "创建暴露 8080 端口映射 80 的 ClusterIP 服务", "yaml_service", "easy",
			"apiVersion: v1\nkind: Service\nmetadata:\n  name: my-svc\nspec:\n  type: ClusterIP\n  selector:\n    app: nginx\n  ports:\n  - port: 8080\n    targetPort: 80",
			[]seedPoint{
				pt("port", "8080", "port", "easy", "Service 对外端口", "text", nil),
				pt("target", "80", "port", "easy", "后端容器端口", "text", nil),
			},
			"ports 中两个端口字段？", "`port`（服务端口）与 `targetPort`（容器端口）", "`port: 8080` / `targetPort: 80`",
			"`port` 是 Service 的 VIP 端口，`targetPort` 是转发到容器的端口。"),
		cfg("k8s-basics-056", "NodePort 服务", "创建固定节点端口 30080 的 NodePort 服务", "yaml_service", "easy",
			"apiVersion: v1\nkind: Service\nmetadata:\n  name: web\nspec:\n  type: NodePort\n  selector:\n    app: web\n  ports:\n  - port: 80\n    targetPort: 80\n    nodePort: 30080",
			[]seedPoint{
				pt("type", "NodePort", "service_type", "easy", "Service 类型", "text", nil),
				pt("nodeport", "30080", "port", "easy", "节点上暴露的端口（30000-32767）", "text", nil),
			},
			"NodePort 的类型与端口？", "`type: NodePort`，`nodePort: 30080`", "`type: NodePort` / `nodePort: 30080`",
			"NodePort 通过节点 IP:30080 访问；端口范围默认 30000-32767。"),
		cfg("k8s-basics-057", "ConfigMap 数据", "创建含数据库地址配置的 ConfigMap", "yaml_configmap", "easy",
			"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app-config\ndata:\n  DB_HOST: db.example.com\n  DB_PORT: \"5432\"",
			[]seedPoint{
				pt("host", "db.example.com", "config", "beginner", "数据库主机配置值", "text", nil),
				pt("port", "\"5432\"", "config", "beginner", "端口值（字符串需要引号）", "text", nil),
			},
			"ConfigMap 数据如何写？", "`data` 下 `key: value`，数值建议加引号", "`DB_HOST: db.example.com` / `DB_PORT: \"5432\"`",
			"ConfigMap.data 存放明文键值；Pod 可通过 env 或 volume 引用。"),
		cfg("k8s-basics-058", "Secret 数据", "创建含 MySQL 密码的 Secret", "yaml_secret", "easy",
			"apiVersion: v1\nkind: Secret\nmetadata:\n  name: db-secret\ntype: Opaque\ndata:\n  password: bXlwYXNz\nstringData:\n  username: root",
			[]seedPoint{
				pt("type", "Opaque", "secret_type", "easy", "Secret 类型（Opaque=不透明）", "text", nil),
				pt("pwd", "bXlwYXNz", "secret_data", "hard", "data 中值需 base64 编码（bXlwYXNz = mypass）", "text", nil),
			},
			"Secret 的 type 与编码？", "`Opaque`；data 值 base64 编码，stringData 明文", "`type: Opaque` / `data.password: bXlwYXNz`",
			"Secret.data 的 value 必须 base64；stringData 可写明文并在创建时自动编码。"),
		cfg("k8s-basics-059", "Ingress 路由", "将 example.com 的 /api 路由到 api 服务", "yaml_ingress", "medium",
			"apiVersion: networking.k8s.io/v1\nkind: Ingress\nmetadata:\n  name: my-ingress\nspec:\n  rules:\n  - host: example.com\n    http:\n      paths:\n      - path: /api\n        pathType: Prefix\n        backend:\n          service:\n            name: api-service\n            port:\n              number: 80",
			[]seedPoint{
				pt("host", "example.com", "host", "easy", "域名", "text", nil),
				pt("path", "/api", "path", "easy", "URL 路径", "text", nil),
			},
			"Ingress 的 host 与 path？", "`host: example.com`，`path: /api`", "`- host: example.com` / `path: /api`",
			"Ingress 按 host + path 将外部请求路由到集群内 Service。"),
		cfg("k8s-basics-060", "PersistentVolumeClaim", "申请 1Gi 的存储", "yaml_pvc", "easy",
			"apiVersion: v1\nkind: PersistentVolumeClaim\nmetadata:\n  name: data-pvc\nspec:\n  accessModes:\n  - ReadWriteOnce\n  resources:\n    requests:\n      storage: 1Gi",
			[]seedPoint{
				pt("mode", "ReadWriteOnce", "access_mode", "easy", "访问模式（单节点读写）", "text", nil),
				pt("storage", "1Gi", "storage", "easy", "申请容量", "text", nil),
			},
			"PVC 的访问模式与容量？", "`ReadWriteOnce`，`storage: 1Gi`", "`- ReadWriteOnce` / `storage: 1Gi`",
			"PVC 声明存储需求，调度时绑定到满足条件的 PV。"),
		cfg("k8s-basics-061", "Pod 挂载 PVC", "把 data-pvc 挂载到容器的 /data 目录", "yaml_volume", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: data\n      mountPath: /data\n  volumes:\n  - name: data\n    persistentVolumeClaim:\n      claimName: data-pvc",
			[]seedPoint{
				pt("mount", "/data", "mount_path", "easy", "容器内挂载路径", "text", nil),
				pt("claim", "data-pvc", "pvc_name", "easy", "引用的 PVC 名称", "text", nil),
			},
			"挂载路径与 PVC 名称？", "`mountPath: /data`，`claimName: data-pvc`", "`mountPath: /data` / `claimName: data-pvc`",
			"容器 volumeMounts 声明挂载点，volumes 声明来源（PVC/emptyDir/configMap 等）。"),
		cfg("k8s-basics-062", "ResourceQuota 配额", "限制命名空间 CPU 总量为 2 核", "yaml_quota", "medium",
			"apiVersion: v1\nkind: ResourceQuota\nmetadata:\n  name: dev-quota\nspec:\n  hard:\n    requests.cpu: \"2\"\n    limits.cpu: \"4\"",
			[]seedPoint{
				pt("req", "requests.cpu: \"2\"", "resource", "medium", "请求量总上限的键", "text", nil),
				pt("lim", "limits.cpu: \"4\"", "resource", "medium", "限制量总上限的键", "text", nil),
			},
			"ResourceQuota 限制的键？", "`requests.cpu` / `limits.cpu`", "`requests.cpu: \"2\"` / `limits.cpu: \"4\"`",
			"ResourceQuota 汇总限制命名空间内所有 Pod 的请求/限制总量。"),
		cfg("k8s-basics-063", "LimitRange 默认值", "为容器设置默认内存限制", "yaml_limitrange", "medium",
			"apiVersion: v1\nkind: LimitRange\nmetadata:\n  name: mem-limit\nspec:\n  limits:\n  - default:\n      memory: 512Mi\n    defaultRequest:\n      memory: 256Mi\n    type: Container",
			[]seedPoint{
				pt("default", "512Mi", "resource", "medium", "未指定时应用的默认限制", "text", nil),
				pt("type", "Container", "limit_type", "medium", "作用对象类型", "text", []string{"Pod"}),
			},
			"LimitRange 的 default 与 type？", "`default.memory: 512Mi`，`type: Container`", "`memory: 512Mi` / `type: Container`",
			"LimitRange 为未显式声明资源的容器注入默认值并约束范围。"),
		cfg("k8s-basics-064", "Job 一次性任务", "创建执行一次的 Job", "yaml_job", "medium",
			"apiVersion: batch/v1\nkind: Job\nmetadata:\n  name: pi\nspec:\n  template:\n    spec:\n      containers:\n      - name: pi\n        image: perl\n        command: [\"perl\", \"-Mbignum=bpi\", \"-wle\", \"print bpi(2000)\"]\n      restartPolicy: Never\n  backoffLimit: 4",
			[]seedPoint{
				pt("api", "batch/v1", "api_version", "easy", "Job 的 API 版本", "text", nil),
				pt("restart", "Never", "restart_policy", "easy", "Job 容器失败时不能重启，任务需整体重试", "text", nil),
			},
			"Job 的 apiVersion 与 restartPolicy？", "`batch/v1`，`restartPolicy: Never`", "`apiVersion: batch/v1` / `restartPolicy: Never`",
			"Job 运行完成即结束；restartPolicy 只能是 Never 或 OnFailure。"),
		cfg("k8s-basics-065", "CronJob 定时任务", "每天凌晨 3 点执行的定时任务", "yaml_cronjob", "medium",
			"apiVersion: batch/v1\nkind: CronJob\nmetadata:\n  name: backup\nspec:\n  schedule: \"0 3 * * *\"\n  jobTemplate:\n    spec:\n      template:\n        spec:\n          containers:\n          - name: backup\n            image: busybox\n            command: [\"sh\", \"-c\", \"echo backup\"]\n          restartPolicy: Never",
			[]seedPoint{
				pt("schedule", "\"0 3 * * *\"", "cron", "easy", "cron 表达式（分 时 日 月 周）", "text", nil),
				pt("api", "batch/v1", "api_version", "easy", "CronJob 的 API 版本", "text", nil),
			},
			"CronJob 的 schedule 与 apiVersion？", "`\"0 3 * * *\"`（每天 3 点），`batch/v1`", "`schedule: \"0 3 * * *\"` / `apiVersion: batch/v1`",
			"`0 3 * * *` = 每天 03:00；CronJob 按 schedule 周期创建 Job。"),
		cfg("k8s-basics-066", "水平自动伸缩 HPA", "按 CPU 在 1-5 副本间伸缩", "yaml_hpa", "medium",
			"apiVersion: autoscaling/v2\nkind: HorizontalPodAutoscaler\nmetadata:\n  name: web-hpa\nspec:\n  scaleTargetRef:\n    apiVersion: apps/v1\n    kind: Deployment\n    name: web\n  minReplicas: 1\n  maxReplicas: 5\n  metrics:\n  - type: Resource\n    resource:\n      name: cpu\n      target:\n        type: Utilization\n        averageUtilization: 70",
			[]seedPoint{
				pt("min", "1", "replicas", "easy", "最小副本数", "text", nil),
				pt("max", "5", "replicas", "easy", "最大副本数", "text", nil),
				pt("util", "70", "utilization", "hard", "目标 CPU 利用率百分比", "text", nil),
			},
			"HPA 的副本范围与目标利用率？", "`minReplicas: 1`，`maxReplicas: 5`，`averageUtilization: 70`", "`minReplicas: 1` / `maxReplicas: 5` / `averageUtilization: 70`",
			"HPA 持续观测指标（如 CPU），在 min/max 之间自动调整副本数。"),
		cfg("k8s-basics-067", "Namespace 与标签", "创建带标签的命名空间", "yaml_namespace", "easy",
			"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: dev\n  labels:\n    env: dev",
			[]seedPoint{
				pt("name", "dev", "metadata", "beginner", "命名空间名称", "text", nil),
				pt("label", "env: dev", "labels", "easy", "命名空间标签", "text", nil),
			},
			"Namespace 的名称与标签？", "`name: dev`，`labels.env: dev`", "`name: dev` / `env: dev`",
			"命名空间用于隔离资源；标签可用于配额或网络策略选择。"),
		cfg("k8s-basics-068", "ServiceAccount 与自动挂载", "创建不自动挂载令牌的 ServiceAccount", "yaml_sa", "medium",
			"apiVersion: v1\nkind: ServiceAccount\nmetadata:\n  name: ci-runner\nautomountServiceAccountToken: false",
			[]seedPoint{
				pt("name", "ci-runner", "metadata", "beginner", "ServiceAccount 名称", "text", nil),
				pt("automount", "false", "config", "medium", "是否自动挂载 API 令牌到 Pod", "text", nil),
			},
			"ServiceAccount 名称与令牌选项？", "`name: ci-runner`，`automountServiceAccountToken: false`", "`name: ci-runner` / `automountServiceAccountToken: false`",
			"ServiceAccount 默认向 Pod 注入 API 令牌；CI 等可关闭自动挂载。"),
		cfg("k8s-basics-069", "Pod 安全上下文", "以非 root 用户运行容器", "yaml_securitycontext", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: secure\nspec:\n  securityContext:\n    runAsNonRoot: true\n  containers:\n  - name: app\n    image: nginx",
			[]seedPoint{
				pt("nonroot", "runAsNonRoot: true", "security", "easy", "强制非 root 运行的字段", "text", nil),
			},
			"强制非 root 的字段？", "`securityContext.runAsNonRoot: true`", "`runAsNonRoot: true`",
			"securityContext 可设置 runAsUser、runAsNonRoot、capabilities、readOnlyRootFilesystem 等。"),
		cfg("k8s-basics-070", "NetworkPolicy 放行", "允许同命名空间 app=web 标签访问 80 端口", "yaml_networkpolicy", "hard",
			"apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n  name: allow-web\nspec:\n  podSelector:\n    matchLabels:\n      app: web\n  policyTypes:\n  - Ingress\n  ingress:\n  - from:\n    - podSelector:\n        matchLabels:\n          app: web\n    ports:\n    - protocol: TCP\n      port: 80",
			[]seedPoint{
				pt("label", "app: web", "labels", "medium", "目标 Pod 的标签", "text", nil),
				pt("port", "80", "port", "easy", "放行的端口", "text", nil),
			},
			"NetworkPolicy 的目标标签与端口？", "`podSelector.matchLabels.app: web`，`port: 80`", "`app: web` / `port: 80`",
			"NetworkPolicy 默认拒绝未放行的流量；需显式允许来源与端口。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, k := range seeds {
		ex, err := buildExercise(toSeedExercise(k))
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
