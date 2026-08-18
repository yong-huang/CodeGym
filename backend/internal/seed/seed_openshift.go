package seed

import (
	"codegym/backend/internal/model"
)

// toOcSeed converts an OpenShift seed (same shape as k8s) to a seedExercise
// under the "openshift" tech stack and the "基础" subcategory.
func toOcSeed(k k8sBasicSeed) seedExercise {
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
		techStack: "openshift", weight: k.weight, category: k.category, subcategory: "基础",
		fullCode: k.code, expected: k.expected, points: points,
		hints: goHint(k.hintL1, k.hintL2, k.hintL3, "正确答案："+k.code), explanation: k.explanation,
	}
}

// OpenShiftBasics returns the 50-question OpenShift basic bank (oc 命令 + YAML).
func OpenShiftBasics() ([]model.Exercise, error) {
	pt := func(id, text, category, weight, hint string) seedPoint {
		return seedPoint{id: id, text: text, category: category, minWeight: weight,
			weightScore: 1, hint: hint, fillType: "text"}
	}
	cmd := func(id, title, desc, category, code string, p seedPoint, l1, l2, l3, expl string) k8sBasicSeed {
		return k8sBasicSeed{id: id, title: title, desc: desc, kind: "command", category: category,
			weight: "easy", code: code, points: []seedPoint{p}, hintL1: l1, hintL2: l2, hintL3: l3,
			explanation: expl}
	}
	cfg := func(id, title, desc, category, weight, code string, points []seedPoint, l1, l2, l3, _l4, expl string) k8sBasicSeed {
		return k8sBasicSeed{id: id, title: title, desc: desc, kind: "config", category: category,
			weight: weight, code: code, points: points, hintL1: l1, hintL2: l2, hintL3: l3,
			explanation: expl}
	}

	seeds := []k8sBasicSeed{
		// ---- oc 命令补全 ----
		cmd("oc-basics-001", "查看当前用户", "查看当前登录的 OpenShift 用户", "oc_command",
			"oc whoami",
			pt("verb", "whoami", "command_verb", "beginner", "查看当前身份的子命令"),
			"查看当前用户的命令", "`oc whoami`", "`oc whoami`",
			"`oc whoami` 显示当前登录用户；`oc whoami -t` 显示 token。"),

		cmd("oc-basics-002", "查看项目列表", "列出所有可访问的项目", "oc_project",
			"oc get projects",
			pt("res", "projects", "resource", "beginner", "项目资源名"),
			"查看项目的资源名", "`oc get projects`", "`oc get projects`",
			"项目（Project）是 OpenShift 对命名空间的一层封装，`get projects` 列出全部。"),

		cmd("oc-basics-003", "查看当前项目", "查看当前使用的项目/命名空间", "oc_project",
			"oc project",
			pt("verb", "project", "command_verb", "easy", "查看/切换项目的子命令"),
			"查看当前项目的命令", "`oc project`", "`oc project`",
			"不带参数 `oc project` 显示当前项目，带参数则切换。"),

		cmd("oc-basics-004", "切换项目", "切换到 my-project 项目", "oc_project",
			"oc project my-project",
			pt("name", "my-project", "project", "easy", "目标项目名"),
			"切换项目的命令", "`oc project <项目名>`", "`oc project my-project`",
			"`oc project <名>` 切换当前上下文的项目，相当于 `kubectl -n`。"),

		cmd("oc-basics-005", "创建项目", "创建名为 demo 的项目", "oc_project",
			"oc new-project demo",
			pt("name", "demo", "project", "beginner", "项目名"),
			"创建项目的命令", "`oc new-project demo`", "`oc new-project demo`",
			"`new-project` 创建独立项目（含配额与角色绑定），是 OpenShift 特有命令。"),

		cmd("oc-basics-006", "删除项目", "删除名为 demo 的项目", "oc_project",
			"oc delete project demo",
			pt("verb", "delete", "command_verb", "medium", "删除资源的动词"),
			"删除项目的命令", "`oc delete project demo`", "`oc delete project demo`",
			"删除项目会连带删除其中所有资源，需谨慎。"),

		cmd("oc-basics-007", "查看 Pod", "列出当前项目下的 Pod", "oc_get",
			"oc get pods",
			pt("res", "pods", "resource", "beginner", "Pod 资源名"),
			"查看 Pod 的命令", "`oc get pods`", "`oc get pods`",
			"`oc get pods` 与 kubectl 一致，列出 Pod。"),

		cmd("oc-basics-008", "查看 DeploymentConfig", "列出 DeploymentConfig", "oc_get",
			"oc get dc",
			pt("res", "dc", "resource", "easy", "DeploymentConfig 缩写"),
			"DeploymentConfig 的缩写", "`dc` = DeploymentConfig", "`oc get dc`",
			"DeploymentConfig（dc）是 OpenShift 的部署对象，支持自动构建/回滚触发。"),

		cmd("oc-basics-009", "查看 Service", "列出 Service", "oc_get",
			"oc get svc",
			pt("res", "svc", "resource", "beginner", "Service 缩写"),
			"Service 的缩写", "`svc` = Service", "`oc get svc`",
			"`svc` 是 Service 的缩写，与 kubectl 一致。"),

		cmd("oc-basics-010", "查看 Route", "列出路由（外部访问入口）", "oc_get",
			"oc get routes",
			pt("res", "routes", "resource", "easy", "Route 资源名"),
			"Route 资源名", "`oc get routes`", "`oc get routes`",
			"Route 是 OpenShift 对外暴露服务的入口，生成可访问的主机名。"),

		cmd("oc-basics-011", "查看 BuildConfig", "列出构建配置", "oc_get",
			"oc get bc",
			pt("res", "bc", "resource", "easy", "BuildConfig 缩写"),
			"BuildConfig 的缩写", "`bc` = BuildConfig", "`oc get bc`",
			"BuildConfig（bc）定义如何从源码构建镜像。"),

		cmd("oc-basics-012", "查看 ImageStream", "列出镜像流", "oc_get",
			"oc get is",
			pt("res", "is", "resource", "easy", "ImageStream 缩写"),
			"ImageStream 的缩写", "`is` = ImageStream", "`oc get is`",
			"ImageStream（is）管理镜像与标签，提供构建/部署的镜像来源。"),

		cmd("oc-basics-013", "应用清单", "用 apply 应用 YAML 清单", "oc_apply",
			"oc apply -f deploymentconfig.yaml",
			pt("flag", "-f deploymentconfig.yaml", "command_flag", "beginner", "指定清单文件"),
			"应用清单的选项", "`-f <文件>`", "`oc apply -f deploymentconfig.yaml`",
			"`oc apply -f` 声明式应用清单，语法与 kubectl 相同。"),

		cmd("oc-basics-014", "从源码建应用", "从 Git 仓库创建应用", "oc_newapp",
			"oc new-app https://github.com/example/demo.git",
			pt("url", "https://github.com/example/demo.git", "source", "easy", "源码仓库地址"),
			"从源码创建应用的命令", "`oc new-app <仓库>`", "`oc new-app https://github.com/example/demo.git`",
			"`new-app` 自动识别语言、创建 BuildConfig/DeploymentConfig/Service 等。"),

		cmd("oc-basics-015", "从镜像建应用", "用镜像创建应用", "oc_newapp",
			"oc new-app --image=nginx:1.25",
			pt("flag", "--image=nginx:1.25", "command_flag", "easy", "指定镜像"),
			"用镜像创建应用的选项", "`--image=<镜像>`", "`oc new-app --image=nginx:1.25`",
			"`new-app --image` 直接以现成镜像部署，不触发源码构建。"),

		cmd("oc-basics-016", "暴露服务", "把 Service 暴露为 Route", "oc_expose",
			"oc expose svc/myapp",
			pt("verb", "expose", "command_verb", "easy", "创建 Route 的动词"),
			"暴露服务为路由的命令", "`oc expose svc/<名>`", "`oc expose svc/myapp`",
			"`oc expose` 为 Service 生成 Route，对外可访问。"),

		cmd("oc-basics-017", "启动构建", "手动触发一次构建", "oc_build",
			"oc start-build myapp",
			pt("verb", "start-build", "command_verb", "easy", "触发构建的命令"),
			"触发构建的命令", "`oc start-build <bc>`", "`oc start-build myapp`",
			"`start-build` 手动触发 BuildConfig 执行一次构建。"),

		cmd("oc-basics-018", "查看构建日志", "实时查看构建日志", "oc_build",
			"oc logs -f bc/myapp",
			pt("flag", "-f", "command_flag", "easy", "实时跟随日志"),
			"跟随日志的选项", "`-f` 实时跟随", "`oc logs -f bc/myapp`",
			"`oc logs -f bc/<名>` 跟踪构建过程日志，排障构建失败用。"),

		cmd("oc-basics-019", "进入容器终端", "打开容器的交互式终端", "oc_debug",
			"oc rsh myapp-pod",
			pt("verb", "rsh", "command_verb", "easy", "进容器执行命令"),
			"进入容器终端的命令", "`oc rsh <pod>`", "`oc rsh myapp-pod`",
			"`oc rsh` 相当于 `kubectl exec -it` 的快捷方式，进入容器 shell。"),

		cmd("oc-basics-020", "复制文件", "在本地与容器之间复制文件", "oc_debug",
			"oc rsync ./local myapp-pod:/tmp",
			pt("verb", "rsync", "command_verb", "medium", "复制文件的命令"),
			"复制文件的命令", "`oc rsync <本地> <pod>:<路径>`", "`oc rsync ./local myapp-pod:/tmp`",
			"`oc rsync` 在本地与容器间同步文件，调试与注入文件常用。"),

		cmd("oc-basics-021", "端口转发", "把服务端口转发到本地", "oc_debug",
			"oc port-forward service/myapp 8080:80",
			pt("ports", "8080:80", "port", "easy", "本地端口:目标端口"),
			"端口转发的写法", "`本地端口:目标端口`", "`oc port-forward service/myapp 8080:80`",
			"`oc port-forward` 把集群内端口转发到本地，便于本地调试。"),

		cmd("oc-basics-022", "查看事件", "查看集群事件", "oc_get",
			"oc get events",
			pt("res", "events", "resource", "easy", "事件资源名"),
			"查看事件的资源名", "`oc get events`", "`oc get events`",
			"`events` 记录调度、镜像拉取、重启等事件，排障重要。"),

		cmd("oc-basics-023", "查看节点", "列出集群节点", "oc_get",
			"oc get nodes",
			pt("res", "nodes", "resource", "beginner", "节点资源名"),
			"查看节点的命令", "`oc get nodes`", "`oc get nodes`",
			"`oc get nodes` 查看集群工作节点。"),

		cmd("oc-basics-024", "查看集群 Operator", "列出集群级 Operator 状态", "oc_get",
			"oc get co",
			pt("res", "co", "resource", "medium", "ClusterOperator 缩写"),
			"ClusterOperator 的缩写", "`co` = ClusterOperator", "`oc get co`",
			"`co` 查看集群各组件的 Operator 健康状态，是 OpenShift 特有的运维视角。"),

		cmd("oc-basics-025", "查看项目内全部资源", "一条命令列出项目内主要资源", "oc_get",
			"oc get all",
			pt("res", "all", "resource", "easy", "聚合资源关键字"),
			"列出全部资源的命令", "`oc get all`", "`oc get all`",
			"`oc get all` 快速列出项目内 Pod/DC/Service/Route 等主要资源。"),

		cmd("oc-basics-026", "扩缩容 DeploymentConfig", "把 myapp 扩到 3 个副本", "oc_scale",
			"oc scale dc/myapp --replicas=3",
			pt("flag", "--replicas=3", "command_flag", "easy", "设置副本数"),
			"设置副本数的选项", "`--replicas=<数>`", "`oc scale dc/myapp --replicas=3`",
			"`oc scale` 调整 DeploymentConfig 副本数。"),

		cmd("oc-basics-027", "按标签筛选", "筛选 app=myapp 的 Pod", "oc_get",
			"oc get pods -l app=myapp",
			pt("flag", "-l app=myapp", "command_flag", "easy", "标签选择器"),
			"标签筛选的选项", "`-l <key>=<value>`", "`oc get pods -l app=myapp`",
			"`-l` 按标签选择器过滤，支持多个键值。"),

		cmd("oc-basics-028", "查看滚动状态", "查看 DeploymentConfig 滚动状态", "oc_rollout",
			"oc rollout status dc/myapp",
			pt("verb", "status", "command_verb", "easy", "查看状态子命令"),
			"查看滚动状态的命令", "`oc rollout status dc/<名>`", "`oc rollout status dc/myapp`",
			"`oc rollout status` 阻塞直到部署完成。"),

		cmd("oc-basics-029", "回滚部署", "回滚 DeploymentConfig 到上一版本", "oc_rollout",
			"oc rollout undo dc/myapp",
			pt("verb", "undo", "command_verb", "easy", "回滚子命令"),
			"回滚的命令", "`oc rollout undo dc/<名>`", "`oc rollout undo dc/myapp`",
			"`oc rollout undo` 回滚到上一版本，`--to-revision` 指定版本。"),

		cmd("oc-basics-030", "编辑资源", "编辑 DeploymentConfig", "oc_edit",
			"oc edit dc/myapp",
			pt("verb", "edit", "command_verb", "easy", "编辑动词"),
			"编辑资源的命令", "`oc edit dc/<名>`", "`oc edit dc/myapp`",
			"`oc edit` 打开编辑器修改资源，保存即生效。"),

		cmd("oc-basics-031", "删除 Pod", "删除指定 Pod", "oc_delete",
			"oc delete pod myapp-abc",
			pt("verb", "delete", "command_verb", "easy", "删除动词"),
			"删除资源的命令", "`oc delete pod <名>`", "`oc delete pod myapp-abc`",
			"`oc delete` 删除资源，DeploymentConfig 会自动重建被删的 Pod。"),

		cmd("oc-basics-032", "打标签", "给应用打标签", "oc_label",
			"oc label pod myapp app=web",
			pt("verb", "label", "command_verb", "medium", "管理标签动词"),
			"管理标签的命令", "`oc label <资源> <key>=<value>`", "`oc label pod myapp app=web`",
			"`oc label` 增改标签，用于选择器与路由。"),

		cmd("oc-basics-033", "查看版本", "查看 oc 与集群版本", "oc_get",
			"oc version",
			pt("verb", "version", "command_verb", "beginner", "查看版本命令"),
			"查看版本的命令", "`oc version`", "`oc version`",
			"`oc version` 显示客户端与集群版本。"),

		cmd("oc-basics-034", "登录集群", "用用户名密码登录", "oc_login",
			"oc login -u admin -p secret https://api.example.com",
			pt("url", "https://api.example.com", "server", "easy", "API 服务器地址"),
			"登录的命令", "`oc login -u admin`", "`oc login -u admin -p secret <api>`",
			"`oc login` 登录集群并写入 kubeconfig；也可用 token 登录。"),

		cmd("oc-basics-035", "查看当前 token", "显示当前用户的服务令牌", "oc_login",
			"oc whoami -t",
			pt("flag", "-t", "command_flag", "medium", "输出 token 选项"),
			"显示 token 的选项", "`-t` 输出 token", "`oc whoami -t`",
			"`oc whoami -t` 输出当前用户 token，用于集成/脚本。"),

		cmd("oc-basics-036", "查看 Secret", "列出 Secret", "oc_get",
			"oc get secrets",
			pt("res", "secrets", "resource", "easy", "Secret 资源名"),
			"Secret 资源名", "`oc get secrets`", "`oc get secrets`",
			"`secrets` 存放敏感信息，Deployment 常以环境变量注入。"),

		cmd("oc-basics-037", "查看 ConfigMap", "列出 ConfigMap", "oc_get",
			"oc get configmaps",
			pt("res", "configmaps", "resource", "easy", "ConfigMap 资源名"),
			"ConfigMap 资源名", "`oc get configmaps`", "`oc get configmaps`",
			"ConfigMap 存非敏感配置，可挂载为文件或注入环境变量。"),

		cmd("oc-basics-038", "查看 PVC", "列出持久卷声明", "oc_get",
			"oc get pvc",
			pt("res", "pvc", "resource", "easy", "PVC 缩写"),
			"PVC 的缩写", "`pvc` = PersistentVolumeClaim", "`oc get pvc`",
			"PVC 请求持久存储，Deployment 中挂载用于有状态应用。"),

		cmd("oc-basics-039", "查看 ServiceAccount", "列出 ServiceAccount", "oc_get",
			"oc get serviceaccounts",
			pt("res", "serviceaccounts", "resource", "easy", "ServiceAccount 资源名"),
			"ServiceAccount 资源名", "`oc get serviceaccounts`", "`oc get serviceaccounts`",
			"ServiceAccount 供容器内进程以特定身份访问集群。"),

		cmd("oc-basics-040", "查看资源配额", "列出配额用量", "oc_get",
			"oc get quota",
			pt("res", "quota", "resource", "medium", "配额资源名"),
			"配额资源名", "`oc get quota`", "`oc get quota`",
			"`quota` 显示项目资源配额及当前用量。"),

		// ---- OpenShift YAML 配置 ----
		cfg("oc-basics-041", "DeploymentConfig 基础", "定义 2 副本的 DeploymentConfig", "yaml_dc", "easy",
			"apiVersion: apps.openshift.io/v1\nkind: DeploymentConfig\nmetadata:\n  name: myapp\nspec:\n  replicas: 2\n  selector:\n    app: myapp\n  template:\n    metadata:\n      labels:\n        app: myapp\n    spec:\n      containers:\n      - name: myapp\n        image: myapp:latest",
			[]seedPoint{
				pt("api", "apps.openshift.io/v1", "api_version", "easy", "DeploymentConfig 的 API 版本"),
				pt("replicas", "2", "replicas", "easy", "副本数"),
			},
			"`apps.openshift.io/v1` / `replicas: 2`", "DeploymentConfig 定义", "属于 openshift apps 组", "`apiVersion: apps.openshift.io/v1`",
			"DeploymentConfig（DC）用 apps.openshift.io/v1，支持触发器自动更新。"),

		cfg("oc-basics-042", "Route 暴露服务", "把 myapp 服务暴露到指定域名", "yaml_route", "easy",
			"apiVersion: route.openshift.io/v1\nkind: Route\nmetadata:\n  name: myapp\nspec:\n  host: myapp.example.com\n  to:\n    kind: Service\n    name: myapp\n  port:\n    targetPort: 8080",
			[]seedPoint{
				pt("host", "myapp.example.com", "host", "easy", "对外域名"),
				pt("svc", "myapp", "service", "easy", "后端 Service 名"),
				pt("port", "8080", "port", "easy", "目标端口"),
			},
			"`myapp.example.com` / `myapp` / `8080`", "Route 定义", "`route.openshift.io/v1`，`to.name` 指 Service", "`host` / `name: myapp` / `targetPort: 8080`",
			"Route 用 route.openshift.io/v1，host 对外的域名，转发到后端 Service 端口。"),

		cfg("oc-basics-043", "BuildConfig 源码构建", "定义从 Git 构建 Source 镜像", "yaml_build", "medium",
			"apiVersion: build.openshift.io/v1\nkind: BuildConfig\nmetadata:\n  name: demo\nspec:\n  source:\n    type: Git\n    git:\n      uri: \"https://github.com/example/demo.git\"\n  strategy:\n    type: Source\n  output:\n    to:\n      kind: ImageStreamTag\n      name: demo:latest",
			[]seedPoint{
				pt("uri", "https://github.com/example/demo.git", "git_uri", "easy", "源码仓库地址"),
				pt("strategy", "Source", "build_strategy", "medium", "构建策略类型"),
			},
			"`https://github.com/example/demo.git` / `Source`", "BuildConfig 定义", "`build.openshift.io/v1`，`strategy.type` 决定构建方式", "`git.uri` / `strategy.type: Source`",
			"BuildConfig 从 Git 拉取源码并构建；Source strategy 产出自定义镜像。"),

		cfg("oc-basics-044", "ImageStream 定义", "创建镜像流并允许按名解析", "yaml_is", "easy",
			"apiVersion: image.openshift.io/v1\nkind: ImageStream\nmetadata:\n  name: demo\nspec:\n  lookupPolicy:\n    local: true",
			[]seedPoint{
				pt("api", "image.openshift.io/v1", "api_version", "easy", "ImageStream 的 API 版本"),
				pt("name", "demo", "metadata", "beginner", "镜像流名称"),
			},
			"`image.openshift.io/v1` / `demo`", "ImageStream 定义", "`image.openshift.io/v1`，`lookupPolicy.local` 允许本地解析", "`apiVersion: image.openshift.io/v1`",
			"ImageStream 跟踪镜像标签，是构建/部署之间的镜像来源。"),

		cfg("oc-basics-045", "Service 定义", "为应用定义 Service", "yaml_service", "easy",
			"apiVersion: v1\nkind: Service\nmetadata:\n  name: myapp\nspec:\n  selector:\n    app: myapp\n  ports:\n  - port: 8080\n    targetPort: 8080",
			[]seedPoint{
				pt("selector", "app: myapp", "selector", "easy", "选中后端 Pod 的标签"),
				pt("port", "8080", "port", "easy", "服务端口"),
			},
			"`app: myapp` / `8080`", "Service 定义", "`selector` 匹配 Pod 标签", "`selector: {app: myapp}`",
			"Service 通过 selector 匹配 Pod，提供稳定的访问入口。"),

		cfg("oc-basics-046", "Route 配置 TLS", "用 edge 终止的 TLS 路由", "yaml_route", "medium",
			"apiVersion: route.openshift.io/v1\nkind: Route\nmetadata:\n  name: secure\nspec:\n  to:\n    kind: Service\n    name: secure-app\n  tls:\n    termination: edge\n    insecureEdgeTerminationPolicy: Redirect",
			[]seedPoint{
				pt("termination", "edge", "tls", "medium", "TLS 终止方式"),
				pt("svc", "secure-app", "service", "easy", "后端 Service 名"),
			},
			"`edge` / `secure-app`", "TLS 路由", "`termination: edge` 在路由层终止 TLS", "`tls.termination: edge`",
			"边缘终止（edge）在 Router 层解密后明文转发到后端，还可强制重定向 https。"),

		cfg("oc-basics-047", "DeploymentConfig 触发器", "配置镜像变更自动部署", "yaml_dc", "medium",
			"apiVersion: apps.openshift.io/v1\nkind: DeploymentConfig\nmetadata:\n  name: myapp\nspec:\n  replicas: 2\n  triggers:\n  - type: ConfigChange\n  - type: ImageChange\n    imageChangeParams:\n      automatic: true\n      containerNames:\n      - myapp\n      from:\n        kind: ImageStreamTag\n        name: myapp:latest\n  selector:\n    app: myapp\n  template:\n    metadata:\n      labels:\n        app: myapp\n    spec:\n      containers:\n      - name: myapp\n        image: myapp:latest",
			[]seedPoint{
				pt("type", "ImageChange", "trigger", "medium", "镜像变更触发器类型"),
				pt("img", "myapp:latest", "image", "medium", "监听镜像流标签"),
			},
			"`ImageChange` / `myapp:latest`", "部署触发器", "`ImageChange` 在镜像更新时自动部署", "`- type: ImageChange`",
			"触发器（triggers）让 DC 在配置或镜像变化时自动滚动部署。"),

		cfg("oc-basics-048", "存活探针", "给 DC 容器配 liveness 探针", "yaml_probe", "medium",
			"apiVersion: apps.openshift.io/v1\nkind: DeploymentConfig\nmetadata:\n  name: web\nspec:\n  replicas: 1\n  selector:\n    app: web\n  template:\n    metadata:\n      labels:\n        app: web\n    spec:\n      containers:\n      - name: web\n        image: nginx\n        livenessProbe:\n          httpGet:\n            path: /healthz\n            port: 80\n          initialDelaySeconds: 5\n          periodSeconds: 10",
			[]seedPoint{
				pt("probe", "livenessProbe", "probe", "medium", "存活探针字段"),
				pt("port", "80", "probe_port", "easy", "探测端口"),
			},
			"`livenessProbe` / `80`", "存活探针", "`httpGet` 探测 /healthz", "`livenessProbe: {httpGet: {path: /healthz}}`",
			"livenessProbe 判断容器存活，失败则重启容器。"),

		cfg("oc-basics-049", "环境变量引用 ConfigMap", "把 ConfigMap 值注入环境变量", "yaml_env", "easy",
			"apiVersion: apps.openshift.io/v1\nkind: DeploymentConfig\nmetadata:\n  name: app\nspec:\n  selector:\n    app: app\n  template:\n    metadata:\n      labels:\n        app: app\n    spec:\n      containers:\n      - name: app\n        image: myapp\n        env:\n        - name: LOG_LEVEL\n          valueFrom:\n            configMapKeyRef:\n              name: app-config\n              key: LOG_LEVEL",
			[]seedPoint{
				pt("cm", "app-config", "configmap", "easy", "ConfigMap 名"),
				pt("key", "LOG_LEVEL", "config", "easy", "取的键"),
			},
			"`app-config` / `LOG_LEVEL`", "env 引用 ConfigMap", "`configMapKeyRef` 注入", "`valueFrom.configMapKeyRef`",
			"通过 configMapKeyRef 把 ConfigMap 的键注入为环境变量。"),

		cfg("oc-basics-050", "PVC 持久存储", "申请 1Gi 存储供应用使用", "yaml_pvc", "easy",
			"apiVersion: v1\nkind: PersistentVolumeClaim\nmetadata:\n  name: data\nspec:\n  accessModes:\n  - ReadWriteOnce\n  resources:\n    requests:\n      storage: 1Gi",
			[]seedPoint{
				pt("mode", "ReadWriteOnce", "access_mode", "easy", "访问模式"),
				pt("storage", "1Gi", "storage", "easy", "申请容量"),
			},
			"`ReadWriteOnce` / `1Gi`", "PVC 定义", "`accessModes` 访问模式，`storage` 容量", "`- ReadWriteOnce` / `storage: 1Gi`",
			"PVC 请求持久存储并与 PV 绑定，之后在 DC 中挂载。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, k := range seeds {
		ex, err := buildExercise(toOcSeed(k))
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
