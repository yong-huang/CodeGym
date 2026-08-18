package seed

import (
	"codegym/backend/internal/model"
)

// toOcDiffSeed converts an OpenShift diff seed (OCP vs Kubernetes) into a
// seedExercise under the "openshift" tech stack and the "对比" subcategory.
func toOcDiffSeed(k k8sBasicSeed) seedExercise {
	ex := toOcSeed(k)
	ex.subcategory = "对比"
	return ex
}

// OpenShiftDiff returns the 30-question OpenShift-vs-Kubernetes bank (对比):
// DeploymentConfig/Routes/BuildConfig/ImageStream/MachineConfig/OLM/SCC 等 OCP 特有对象。
func OpenShiftDiff() ([]model.Exercise, error) {
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
		// ---- OCP 特有对象 / 命令 ----
		cmd("oc-diff-001", "DeploymentConfig 与 Deployment", "用 oc 查看 DeploymentConfig（K8s 没有的对象）", "dc",
			"oc get dc",
			pt("res", "dc", "resource", "beginner", "DeploymentConfig 缩写"),
			"DeploymentConfig 的缩写", "`dc` = DeploymentConfig", "`oc get dc`",
			"DeploymentConfig 是 OCP 对 Deployment 的扩展，支持触发器自动部署。"),

		cmd("oc-diff-002", "ClusterOperator", "查看集群级 Operator 状态", "operator",
			"oc get co",
			pt("res", "co", "resource", "easy", "ClusterOperator 缩写"),
			"ClusterOperator 的缩写", "`co` = ClusterOperator", "`oc get co`",
			"ClusterOperator 管理集群组件，K8s 没有此对象；OCP 用它做组件级自愈。"),

		cmd("oc-diff-003", "MachineConfig", "查看 MachineConfig（OCP 节点配置）", "machine",
			"oc get mc",
			pt("res", "mc", "resource", "medium", "MachineConfig 缩写"),
			"MachineConfig 的缩写", "`mc` = MachineConfig", "`oc get mc`",
			"MachineConfig 给节点（RHCOS）注入系统配置，是 OCP 节点管理核心。"),

		cmd("oc-diff-004", "MachineConfigPool", "查看 MachineConfigPool", "machine",
			"oc get mcp",
			pt("res", "mcp", "resource", "medium", "MachineConfigPool 缩写"),
			"MachineConfigPool 的缩写", "`mcp` = MachineConfigPool", "`oc get mcp`",
			"MachineConfigPool 把节点分组，滚动应用 MachineConfig，类似控制面/工作节点分组。"),

		cmd("oc-diff-005", "MachineSet", "查看 MachineSet（Machine API）", "machine",
			"oc get machineset",
			pt("res", "machineset", "resource", "medium", "MachineSet 资源名"),
			"MachineSet 资源名", "`oc get machineset`", "`oc get machineset`",
			"MachineSet 通过 Machine API 管理云上节点生命周期，K8s 仅靠 kubelet+云驱动。"),

		cmd("oc-diff-006", "Machine", "查看 Machine（节点底层对象）", "machine",
			"oc get machines",
			pt("res", "machines", "resource", "medium", "Machine 资源名"),
			"Machine 资源名", "`oc get machines`", "`oc get machines`",
			"Machine 是 Machine API 对底层实例的抽象；OCP 节点增删都经 Machine。"),

		cmd("oc-diff-007", "ClusterVersion", "查看集群升级状态", "upgrade",
			"oc get clusterversion",
			pt("res", "clusterversion", "resource", "medium", "集群版本对象"),
			"集群版本资源名", "`oc get clusterversion`", "`oc get clusterversion`",
			"ClusterVersion 记录 OCP 版本与升级状态，K8s 无对应集群级升级对象。"),

		cmd("oc-diff-008", "OLM Subscription", "查看 Operator 订阅", "olm",
			"oc get subscription",
			pt("res", "subscription", "resource", "hard", "OLM 订阅资源名"),
			"OLM 订阅资源名", "`oc get subscription`", "`oc get subscription`",
			"Subscription 是 OLM 的订阅，把 Operator 从 catalog 安装到集群。"),

		cmd("oc-diff-009", "ClusterServiceVersion", "查看 CSV（Operator 打包）", "olm",
			"oc get csv",
			pt("res", "csv", "resource", "hard", "CSV 缩写"),
			"ClusterServiceVersion 的缩写", "`csv` = ClusterServiceVersion", "`oc get csv`",
			"CSV 描述一个 Operator 的元数据与安装，是 OLM 的核心对象。"),

		cmd("oc-diff-010", "PackageManifest", "查看 OperatorHub 可安装的包", "olm",
			"oc get packagemanifest",
			pt("res", "packagemanifest", "resource", "hard", "包清单资源名"),
			"可安装 Operator 的资源名", "`oc get packagemanifest`", "`oc get packagemanifest`",
			"PackageManifest 来自 OperatorHub，列出可订阅安装的 Operator。"),

		cmd("oc-diff-011", "Route 与 Ingress", "用 oc 查看 Route（对应 K8s Ingress）", "route",
			"oc get routes",
			pt("res", "routes", "resource", "beginner", "Route 资源名"),
			"Route 资源名", "`oc get routes`", "`oc get routes`",
			"Route 是 OCP 对 Ingress 的增强：自带 TLS 终止、权重路由等。"),

		cmd("oc-diff-012", "SecurityContextConstraints", "查看 SCC（安全上下文约束）", "scc",
			"oc get scc",
			pt("res", "scc", "resource", "hard", "SCC 缩写"),
			"SCC 的缩写", "`scc` = SecurityContextConstraints", "`oc get scc`",
			"SCC 取代了废弃的 PodSecurityPolicy，控制 Pod 能以什么身份运行。"),

		cmd("oc-diff-013", "ImageStream", "查看镜像流", "imagestream",
			"oc get is",
			pt("res", "is", "resource", "easy", "ImageStream 缩写"),
			"ImageStream 的缩写", "`is` = ImageStream", "`oc get is`",
			"ImageStream 是 OCP 的镜像抽象（K8s 无），连接构建与部署的镜像来源。"),

		cmd("oc-diff-014", "BuildConfig", "查看构建配置", "build",
			"oc get bc",
			pt("res", "bc", "resource", "easy", "BuildConfig 缩写"),
			"BuildConfig 的缩写", "`bc` = BuildConfig", "`oc get bc`",
			"BuildConfig 定义从源码构建镜像的流程，K8s 本身没有内置构建能力。"),

		cmd("oc-diff-015", "集群升级", "用 oc adm 触发集群升级", "upgrade",
			"oc adm upgrade",
			pt("verb", "adm upgrade", "command_verb", "hard", "管理员升级命令"),
			"集群升级的命令", "`oc adm upgrade`", "`oc adm upgrade`",
			"`oc adm upgrade` 管理 OCP 集群升级（K8s 用 kubeadm/托管平台）。"),

		cmd("oc-diff-016", "批准 CSR", "批准待处理的证书请求", "admin",
			"oc adm certificate approve node-abc",
			pt("verb", "certificate approve", "command_verb", "hard", "批准证书子命令"),
			"批准 CSR 的命令", "`oc adm certificate approve <名>`", "`oc adm certificate approve node-abc`",
			"节点加入集群时需 `oc adm certificate approve` 批准其 CSR（K8s 用 kubectl）。"),

		cmd("oc-diff-017", "节点维护", "用 oc adm 驱逐节点 Pod", "admin",
			"oc adm drain node-1 --ignore-daemonsets",
			pt("flag", "--ignore-daemonsets", "command_flag", "medium", "忽略 DaemonSet"),
			"驱逐节点 Pod 的命令", "`oc adm drain`", "`oc adm drain node-1 --ignore-daemonsets`",
			"`oc adm drain` 与 kubectl 类似，是 OCP 的管理员级节点维护入口。"),

		cmd("oc-diff-018", "查看用户", "查看集群用户", "auth",
			"oc get users",
			pt("res", "users", "resource", "hard", "用户资源名"),
			"用户资源名", "`oc get users`", "`oc get users`",
			"OCP 用 OAuth/身份提供方管理用户，`oc get users` 是 K8s 没有的用户对象。"),

		cmd("oc-diff-019", "系统项目", "查看项目（对应命名空间）", "project",
			"oc get projects",
			pt("res", "projects", "resource", "easy", "项目资源名"),
			"项目资源名", "`oc get projects`", "`oc get projects`",
			"Project 是 OCP 对 Namespace 的封装，自带配额与访问控制，还有 openshift-* 系统项目。"),

		cmd("oc-diff-020", "模板", "查看模板（OCP 版 Helm）", "template",
			"oc get templates",
			pt("res", "templates", "resource", "hard", "模板资源名"),
			"模板资源名", "`oc get templates`", "`oc get templates`",
			"Template 是 OCP 的参数化资源模板，早于 Helm；Helm 也受支持。"),

		cmd("oc-diff-021", "查看 Operator 列表", "查看已安装的 Operator", "operator",
			"oc get operators",
			pt("res", "operators", "resource", "hard", "Operator 资源名"),
			"查看 Operator 的命令", "`oc get operators`", "`oc get operators`",
			"OCP 自带 OLM，`oc get operators` 列出集群中安装的 Operator。"),

		cmd("oc-diff-022", "集群网络配置", "查看 OpenShift 集群网络", "network",
			"oc get network cluster",
			pt("res", "network", "resource", "hard", "网络配置资源名"),
			"集群网络资源名", "`oc get network cluster`", "`oc get network cluster`",
			"`oc get network cluster` 查看 OVN-Kubernetes/OpenShiftSDN 等网络插件配置。"),

		// ---- YAML 配置 ----
		cfg("oc-diff-023", "Route 边缘 TLS", "用 edge 终止的 Route（对比 Ingress TLS）", "route",
			"apiVersion: route.openshift.io/v1\nkind: Route\nmetadata:\n  name: web\nspec:\n  to:\n    kind: Service\n    name: web\n  tls:\n    termination: edge\n    insecureEdgeTerminationPolicy: Redirect",
			[]seedPoint{
				pt("term", "edge", "tls", "hard", "TLS 终止方式"),
				pt("svc", "web", "service", "easy", "后端 Service 名"),
			},
			"`edge` / `web`", "Route vs Ingress", "Route 的 TLS 在 spec.tls 配置，edge 在路由层终止", "`termination: edge`",
			"Route 原生支持 edge/passthrough/reencrypt 三种 TLS，比 K8s Ingress 更直接。"),

		cfg("oc-diff-024", "DeploymentConfig 定义", "用 DC 定义应用（对比 Deployment）", "dc",
			"apiVersion: apps.openshift.io/v1\nkind: DeploymentConfig\nmetadata:\n  name: web\nspec:\n  replicas: 2\n  selector:\n    app: web\n  template:\n    metadata:\n      labels:\n        app: web\n    spec:\n      containers:\n      - name: web\n        image: web:latest",
			[]seedPoint{
				pt("api", "apps.openshift.io/v1", "api_version", "easy", "DC 的 API 版本"),
				pt("replicas", "2", "replicas", "easy", "副本数"),
			},
			"`apps.openshift.io/v1` / `2`", "DeploymentConfig 与 Deployment", "DC 用 apps.openshift.io/v1，Deployment 用 apps/v1", "`apiVersion: apps.openshift.io/v1`",
			"两者都管副本，但 DC 额外支持触发器与自动回滚，是 OCP 的增强部署对象。"),

		cfg("oc-diff-025", "镜像变更触发器", "DC 在镜像更新时自动部署", "dc",
			"apiVersion: apps.openshift.io/v1\nkind: DeploymentConfig\nmetadata:\n  name: web\nspec:\n  replicas: 1\n  triggers:\n  - type: ImageChange\n    imageChangeParams:\n      automatic: true\n      containerNames:\n      - web\n      from:\n        kind: ImageStreamTag\n        name: web:latest\n  selector:\n    app: web\n  template:\n    metadata:\n      labels:\n        app: web\n    spec:\n      containers:\n      - name: web\n        image: web:latest",
			[]seedPoint{
				pt("type", "ImageChange", "trigger", "medium", "镜像变更触发器"),
				pt("from", "web:latest", "image", "medium", "监听镜像流标签"),
			},
			"`ImageChange` / `web:latest`", "DC 触发器", "镜像更新自动触发滚动部署，Deployment 需手动改 image", "`type: ImageChange`",
			"DC 的 ImageChange 触发器在 ImageStream 更新时自动部署，K8s Deployment 需额外手段。"),

		cfg("oc-diff-026", "构建输出镜像流", "BuildConfig 输出到 ImageStreamTag", "build",
			"apiVersion: build.openshift.io/v1\nkind: BuildConfig\nmetadata:\n  name: web\nspec:\n  source:\n    type: Git\n    git:\n      uri: \"https://github.com/example/web.git\"\n  strategy:\n    type: Source\n  output:\n    to:\n      kind: ImageStreamTag\n      name: web:latest",
			[]seedPoint{
				pt("uri", "https://github.com/example/web.git", "git_uri", "easy", "源码仓库地址"),
				pt("out", "web:latest", "image", "medium", "构建产物写入的镜像流标签"),
			},
			"`https://github.com/example/web.git` / `web:latest`", "构建流水线", "OCP 源码构建输出到 ImageStreamTag", "`output.to.name: web:latest`",
			"OCP 内置 Source-to-Image 构建流水线，构建产物写入 ImageStreamTag 供部署。"),

		cfg("oc-diff-027", "MachineConfig 配置", "给节点注入内核参数", "machine",
			"apiVersion: machineconfiguration.openshift.io/v1\nkind: MachineConfig\nmetadata:\n  name: 99-worker-sysctl\n  labels:\n    machineconfiguration.openshift.io/role: worker\nspec:\n  config:\n    ignition:\n      version: 3.2.0\n    storage:\n      files:\n      - path: /etc/sysctl.d/99-fs.conf\n        contents:\n          source: data:,%20fs.file-max%20%3D%2050000\n",
			[]seedPoint{
				pt("api", "machineconfiguration.openshift.io/v1", "api_version", "hard", "MachineConfig 的 API 组"),
				pt("role", "worker", "label", "medium", "作用到哪类节点（worker/master）"),
			},
			"`machineconfiguration.openshift.io/v1` / `worker`", "MachineConfig", "用 ignition 配置节点，K8s 无此对象", "`machineconfiguration.openshift.io/v1`",
			"MachineConfig 用 Ignition 给 RHCOS 节点下发系统级配置，这是 OCP 节点管理核心。"),

		cfg("oc-diff-028", "SCC 安全约束", "允许特权容器", "scc",
			"apiVersion: security.openshift.io/v1\nkind: SecurityContextConstraints\nmetadata:\n  name: privileged\nallowPrivilegedContainer: true\nallowedCapabilities:\n- \"*\"\nrunAsUser:\n  type: RunAsAny",
			[]seedPoint{
				pt("api", "security.openshift.io/v1", "api_version", "hard", "SCC 的 API 组"),
				pt("priv", "allowPrivilegedContainer: true", "security", "hard", "允许特权容器"),
			},
			"`security.openshift.io/v1` / `true`", "SCC vs PSP", "SCC 用 security.openshift.io/v1，取代废弃的 PSP", "`allowPrivilegedContainer: true`",
			"SCC 决定 Pod 能否以特权/root 运行，是 OCP 的安全扩展（K8s 已弃用 PSP）。"),

		cfg("oc-diff-029", "ImageStream 定义", "定义镜像流", "imagestream",
			"apiVersion: image.openshift.io/v1\nkind: ImageStream\nmetadata:\n  name: web\nspec:\n  lookupPolicy:\n    local: true",
			[]seedPoint{
				pt("api", "image.openshift.io/v1", "api_version", "easy", "ImageStream 的 API 组"),
				pt("name", "web", "metadata", "beginner", "镜像流名"),
			},
			"`image.openshift.io/v1` / `web`", "ImageStream", "K8s 没有镜像流，只有镜像引用", "`apiVersion: image.openshift.io/v1`",
			"ImageStream 用 image.openshift.io/v1，跟踪镜像标签供构建/部署使用。"),

		cfg("oc-diff-030", "集群网络配置", "查看/配置 OVN 集群网络", "network",
			"apiVersion: config.openshift.io/v1\nkind: Network\nmetadata:\n  name: cluster\nspec:\n  networkType: OVNKubernetes\n",
			[]seedPoint{
				pt("api", "config.openshift.io/v1", "api_version", "hard", "集群配置 API 组"),
				pt("net", "OVNKubernetes", "network", "hard", "网络插件类型"),
			},
			"`config.openshift.io/v1` / `OVNKubernetes`", "集群网络", "Network 对象配置网络插件，K8s 无此集群级对象", "`networkType: OVNKubernetes`",
			"OCP 用 `Network` 集群对象配置 OVN-Kubernetes 网络插件，K8s 由 CNI 提供者自行管理。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, k := range seeds {
		ex, err := buildExercise(toOcDiffSeed(k))
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
