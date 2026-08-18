package seed

import (
	"codegym/backend/internal/model"
)

// K8sResourceConfig returns the Kubernetes "资源配置" series (30 questions):
// YAML configs for Secret, ConfigMap, Volumes, Probes, resources, etc.
func K8sResourceConfig() ([]model.Exercise, error) {
	pt := func(id, text, category, weight, hint string) seedPoint {
		return seedPoint{id: id, text: text, category: category, minWeight: weight,
			weightScore: 1, hint: hint, fillType: "text"}
	}
	cfg := func(id, title, desc, category, weight, code string, points []seedPoint, l4, l1, l2, l3, expl string) seedExercise {
		return seedExercise{
			id: id, title: title, description: desc, exerciseType: model.ExerciseTypeConfig,
			techStack: "kubernetes", weight: weight, category: category, subcategory: "资源配置",
			fullCode: code, points: points,
			hints: goHint(l1, l2, l3, l4), explanation: expl,
		}
	}

	seeds := []seedExercise{
		cfg("k8s-res-001", "创建 Secret 数据", "创建 Opaque 类型 Secret，data 中存 base64 值", "secret", "easy",
			"apiVersion: v1\nkind: Secret\nmetadata:\n  name: db-creds\ntype: Opaque\ndata:\n  username: YWRtaW4=\n  password: c2VjcmV0",
			[]seedPoint{
				pt("type", "Opaque", "secret_type", "easy", "Secret 的 type（Opaque/kubernetes.io/tls 等）"),
				pt("pwd", "c2VjcmV0", "secret_data", "medium", "data 值需 base64 编码（c2VjcmV0 = secret）"),
			},
			"`Opaque` / `c2VjcmV0`", "Secret 的 type 与编码", "data 的值必须 base64；Opaque 是通用类型", "`type: Opaque` / `password: c2VjcmV0`",
			"Secret.data 的值必须 base64 编码；Opaque 表示不透明字节数据。"),

		cfg("k8s-res-002", "Secret 明文 stringData", "用 stringData 直接写明文并自动编码", "secret", "easy",
			"apiVersion: v1\nkind: Secret\nmetadata:\n  name: db-creds\ntype: Opaque\nstringData:\n  username: admin\n  password: secret",
			[]seedPoint{
				pt("k", "username", "config", "beginner", "Secret 的键名"),
				pt("val", "secret", "config", "beginner", "明文密码值"),
			},
			"`username` / `secret`", "stringData 明文键值", "stringData 里不用 base64，自动编码", "`username: admin` / `password: secret`",
			"`stringData` 允许明文，创建时自动编码；`data` 需手写 base64。"),

		cfg("k8s-res-003", "创建 ConfigMap", "创建含两个配置键的 ConfigMap", "configmap", "easy",
			"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app-config\ndata:\n  LOG_LEVEL: info\n  RETRY: \"3\"",
			[]seedPoint{
				pt("level", "info", "config", "beginner", "日志级别值"),
				pt("retry", "3", "config", "beginner", "重试次数（字符串需引号）"),
			},
			"`info` / `\"3\"`", "ConfigMap 的键值", "data 存明文；数字建议加引号", "`LOG_LEVEL: info` / `RETRY: \"3\"`",
			"ConfigMap.data 存非敏感配置，Pod 通过 env 或 volume 引用。"),

		cfg("k8s-res-004", "env 引用 ConfigMap 单项", "用 configMapKeyRef 注入单个配置", "env", "easy",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: web\nspec:\n  containers:\n  - name: web\n    image: nginx\n    env:\n    - name: LOG_LEVEL\n      valueFrom:\n        configMapKeyRef:\n          name: app-config\n          key: LOG_LEVEL",
			[]seedPoint{
				pt("cmname", "app-config", "configmap", "easy", "引用的 ConfigMap 名称"),
				pt("key", "LOG_LEVEL", "config", "easy", "取哪个键"),
			},
			"`app-config` / `LOG_LEVEL`", "env 单项取自 ConfigMap", "`valueFrom.configMapKeyRef` 引用", "`name: app-config` / `key: LOG_LEVEL`",
			"`configMapKeyRef` 把某个 ConfigMap 的键注入为环境变量。"),

		cfg("k8s-res-005", "env 引用 Secret 单项", "用 secretKeyRef 注入密码环境变量", "env", "easy",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: myapp\n    env:\n    - name: DB_PASSWORD\n      valueFrom:\n        secretKeyRef:\n          name: db-creds\n          key: password",
			[]seedPoint{
				pt("secret", "db-creds", "secret", "easy", "引用的 Secret 名称"),
				pt("envname", "DB_PASSWORD", "config", "easy", "容器内环境变量名"),
			},
			"`db-creds` / `DB_PASSWORD`", "env 取自 Secret", "`valueFrom.secretKeyRef` 引用 Secret 键", "`name: db-creds` / `key: password`",
			"`secretKeyRef` 把 Secret 的 value 注入环境变量（解密为明文）。"),

		cfg("k8s-res-006", "envFrom 引用整个 ConfigMap", "把整个 ConfigMap 作为环境变量注入", "env", "easy",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: myapp\n    envFrom:\n    - configMapRef:\n        name: app-config",
			[]seedPoint{
				pt("name", "app-config", "configmap", "easy", "整表引用的 ConfigMap 名"),
			},
			"`app-config`", "整表注入 ConfigMap", "`envFrom` 把每个键变成环境变量", "`envFrom: - configMapRef: name: app-config`",
			"`envFrom.configMapRef` 将 ConfigMap 的所有键一次性注入为环境变量。"),

		cfg("k8s-res-007", "ConfigMap 挂载为文件", "把 ConfigMap 挂到 /config 作为文件", "volume", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: config\n      mountPath: /config\n  volumes:\n  - name: config\n    configMap:\n      name: app-config",
			[]seedPoint{
				pt("mount", "/config", "mount_path", "easy", "挂载路径"),
				pt("cm", "app-config", "configmap", "easy", "来源 ConfigMap 名"),
			},
			"`/config` / `app-config`", "ConfigMap 卷挂载", "`volumes.configMap` 定义来源，`volumeMounts` 挂载", "`mountPath: /config` / `name: app-config`",
			"把 ConfigMap 作为卷挂载后，每个键成为挂载目录下的一个文件。"),

		cfg("k8s-res-008", "Secret 挂载为文件", "把 Secret 挂到 /etc/secret 作为文件", "volume", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: creds\n      mountPath: /etc/secret\n      readOnly: true\n  volumes:\n  - name: creds\n    secret:\n      secretName: db-creds",
			[]seedPoint{
				pt("mount", "/etc/secret", "mount_path", "easy", "挂载路径"),
				pt("sec", "db-creds", "secret", "easy", "来源 Secret 名"),
			},
			"`/etc/secret` / `db-creds`", "Secret 卷挂载", "`volumes.secret.secretName` 指定 Secret", "`mountPath: /etc/secret` / `secretName: db-creds`",
			"Secret 卷挂载后每个键成为文件，值以 base64 解密写入。"),

		cfg("k8s-res-009", "emptyDir 临时卷", "创建一个空临时卷供容器写入", "volume", "easy",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: cache\nspec:\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: scratch\n      mountPath: /tmp/cache\n  volumes:\n  - name: scratch\n    emptyDir: {}",
			[]seedPoint{
				pt("mount", "/tmp/cache", "mount_path", "easy", "临时目录挂载路径"),
				pt("vol", "emptyDir: {}", "volume", "easy", "空卷类型"),
			},
			"`/tmp/cache` / `- name: scratch; emptyDir: {}`", "emptyDir 临时卷", "`emptyDir` 在 Pod 生命期内存在", "`emptyDir: {}`",
			"emptyDir 是 Pod 内的临时空目录，Pod 删除即清空。"),

		cfg("k8s-res-010", "emptyDir 内存介质", "用内存作为 emptyDir 介质（tmpfs）", "volume", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: cache\nspec:\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: mem\n      mountPath: /dev/shm\n  volumes:\n  - name: mem\n    emptyDir:\n      medium: Memory",
			[]seedPoint{
				pt("medium", "Memory", "volume", "medium", "使用内存 tmpfs 的介质值"),
			},
			"`Memory`", "内存介质 emptyDir", "`emptyDir.medium: Memory` 用 tmpfs", "`medium: Memory`",
			"`medium: Memory` 让 emptyDir 使用内存（tmpfs），性能更高但不持久。"),

		cfg("k8s-res-011", "hostPath 卷", "挂载宿主机目录到容器", "volume", "hard",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: host\nspec:\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: hostdir\n      mountPath: /data\n  volumes:\n  - name: hostdir\n    hostPath:\n      path: /var/log\n      type: Directory",
			[]seedPoint{
				pt("path", "/var/log", "host_path", "easy", "宿主机上的目录路径"),
				pt("type", "Directory", "volume", "hard", "hostPath 检查类型"),
			},
			"`/var/log` / `Directory`", "宿主机目录挂载", "`hostPath` 直连节点文件系统", "`path: /var/log` / `type: Directory`",
			"hostPath 直接把节点上的目录/文件挂进容器，适合 DaemonSet 访问节点日志。"),

		cfg("k8s-res-012", "PVC 与 Pod 挂载", "申请 1Gi 存储并挂载到 /data", "storage", "easy",
			"apiVersion: v1\nkind: PersistentVolumeClaim\nmetadata:\n  name: data\nspec:\n  accessModes:\n  - ReadWriteOnce\n  resources:\n    requests:\n      storage: 1Gi\n---\napiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: data\n      mountPath: /data\n  volumes:\n  - name: data\n    persistentVolumeClaim:\n      claimName: data",
			[]seedPoint{
				pt("storage", "1Gi", "storage", "easy", "申请容量"),
				pt("claim", "claimName: data", "pvc", "easy", "引用 PVC 名称"),
			},
			"`1Gi` / `claimName: data`", "PVC 声明与挂载", "先建 PVC，Pod 用 `persistentVolumeClaim.claimName` 引用", "`storage: 1Gi` / `claimName: data`",
			"PVC 声明存储需求并绑定 PV，Pod 通过同名 volume 挂载 PVC。"),

		cfg("k8s-res-013", "StorageClass 定义", "定义一个本地存储类", "storage", "medium",
			"apiVersion: storage.k8s.io/v1\nkind: StorageClass\nmetadata:\n  name: fast\nprovisioner: kubernetes.io/aws-ebs\nparameters:\n  type: gp3\nreclaimPolicy: Delete\nvolumeBindingMode: WaitForFirstConsumer",
			[]seedPoint{
				pt("api", "storage.k8s.io/v1", "api_version", "easy", "StorageClass 的 API 版本"),
				pt("provisioner", "kubernetes.io/aws-ebs", "provisioner", "medium", "卷供应器"),
				pt("reclaim", "Delete", "reclaim_policy", "medium", "删除时回收策略"),
			},
			"`storage.k8s.io/v1` / `kubernetes.io/aws-ebs` / `Delete`",
			"StorageClass 定义动态供应", "`provisioner` 指定存储插件，`reclaimPolicy` 设回收", "`provisioner: kubernetes.io/aws-ebs`",
			"StorageClass 让 PVC 自动创建 PV；reclaimPolicy 控制 PV 释放后的处理。"),

		cfg("k8s-res-014", "livenessProbe 存活探针", "用 httpGet 定时探测容器是否存活", "probe", "hard",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: web\nspec:\n  containers:\n  - name: web\n    image: nginx\n    livenessProbe:\n      httpGet:\n        path: /healthz\n        port: 80\n      initialDelaySeconds: 5\n      periodSeconds: 10",
			[]seedPoint{
				pt("port", "80", "probe_port", "easy", "探测端口"),
				pt("period", "10", "probe", "hard", "探测周期(秒)"),
			},
			"`80` / `10`", "存活探针", "`livenessProbe` 失败则重启容器", "`httpGet: {path: /healthz, port: 80}`",
			"livenessProbe 判断容器是否还活着，失败 Kubernetes 会重启容器。"),

		cfg("k8s-res-015", "readinessProbe 就绪探针", "用 tcpSocket 判断服务是否就绪", "probe", "hard",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: myapp\n    readinessProbe:\n      tcpSocket:\n        port: 8080\n      initialDelaySeconds: 3\n      periodSeconds: 5",
			[]seedPoint{
				pt("port", "8080", "probe_port", "easy", "探测端口"),
				pt("probe", "tcpSocket", "probe", "hard", "TCP 连接探测类型"),
			},
			"`8080` / `tcpSocket`", "就绪探针", "`readinessProbe` 就绪前不接收流量", "`tcpSocket: {port: 8080}`",
			"readinessProbe 通过后 Pod 才被加入 Service 端点接收流量。"),

		cfg("k8s-res-016", "startupProbe 启动探针", "给慢启动容器配置启动探测", "probe", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: slow\nspec:\n  containers:\n  - name: slow\n    image: myapp\n    startupProbe:\n      tcpSocket:\n        port: 8080\n      failureThreshold: 30\n      periodSeconds: 10",
			[]seedPoint{
				pt("threshold", "30", "probe", "medium", "失败多少次算启动失败"),
			},
			"`30`", "启动探针", "`startupProbe` 成功前禁用其它探针", "`failureThreshold: 30`",
			"startupProbe 给慢启动应用充足时间，成功后交由 liveness/readiness 接管。"),

		cfg("k8s-res-017", "镜像拉取策略", "强制每次拉取最新镜像", "image", "easy",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: myapp:latest\n    imagePullPolicy: Always",
			[]seedPoint{
				pt("policy", "Always", "image_pull_policy", "easy", "每次拉取的策略值"),
			},
			"`Always`", "镜像拉取策略", "Always/Never/IfNotPresent", "`imagePullPolicy: Always`",
			"imagePullPolicy 控制拉取时机；latest 标签默认 Always。"),

		cfg("k8s-res-018", "容器资源请求", "声明 CPU 100m / 内存 64Mi 请求", "resources", "easy",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: nginx\n    resources:\n      requests:\n        cpu: 100m\n        memory: 64Mi",
			[]seedPoint{
				pt("cpu", "100m", "resource", "easy", "CPU 请求（100m = 0.1 核）"),
				pt("mem", "64Mi", "resource", "easy", "内存请求"),
			},
			"`100m` / `64Mi`", "资源 requests", "`requests` 是调度保证，`limits` 是硬上限", "`cpu: 100m` / `memory: 64Mi`",
			"requests 声明所需最小资源，调度器据此放置 Pod。"),

		cfg("k8s-res-019", "Pod 使用 ServiceAccount", "为 Pod 指定 ServiceAccount", "sa", "easy",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  serviceAccountName: app-runner\n  containers:\n  - name: app\n    image: nginx",
			[]seedPoint{
				pt("sa", "app-runner", "serviceaccount", "easy", "ServiceAccount 名称"),
			},
			"`app-runner`", "指定 ServiceAccount", "`spec.serviceAccountName` 声明身份", "`serviceAccountName: app-runner`",
			"Pod 通过 serviceAccountName 以指定身份访问 API；默认 default。"),

		cfg("k8s-res-020", "nodeSelector 调度", "只调度到 gpu=true 标签的节点", "scheduling", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: gpu\nspec:\n  nodeSelector:\n    gpu: \"true\"\n  containers:\n  - name: app\n    image: nginx",
			[]seedPoint{
				pt("key", "gpu", "label", "medium", "节点标签键"),
			},
			"`gpu`", "按标签调度", "`nodeSelector` 匹配节点标签键值", "`nodeSelector: {gpu: \"true\"}`",
			"nodeSelector 把 Pod 调度到带指定标签的节点上。"),

		cfg("k8s-res-021", "容忍调度", "容忍 NoSchedule 污点让 Pod 调度到特殊节点", "scheduling", "hard",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: util\nspec:\n  tolerations:\n  - key: app\n    operator: Exists\n    effect: NoSchedule\n  containers:\n  - name: app\n    image: nginx",
			[]seedPoint{
				pt("effect", "NoSchedule", "taint_effect", "hard", "污点效果"),
			},
			"`NoSchedule`", "容忍污点", "`tolerations` 匹配节点污点的 key/effect", "`effect: NoSchedule`",
			"tolerations 允许 Pod 被调度到带对应污点的节点。"),

		cfg("k8s-res-022", "节点亲和性", "用 requiredDuringScheduling 亲和指定节点", "scheduling", "hard",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  affinity:\n    nodeAffinity:\n      requiredDuringSchedulingIgnoredDuringExecution:\n        nodeSelectorTerms:\n        - matchExpressions:\n          - key: zone\n            operator: In\n            values:\n            - us-east-1\n  containers:\n  - name: app\n    image: nginx",
			[]seedPoint{
				pt("key", "zone", "label", "hard", "节点标签键"),
				pt("value", "us-east-1", "label", "medium", "期望的标签值"),
			},
			"`zone` / `us-east-1`", "节点亲和性", "`nodeAffinity` 用表达式选择节点", "`key: zone` / `values: [us-east-1]`",
			"nodeAffinity 比 nodeSelector 更灵活，可处理 In/NotIn 等操作符。"),

		cfg("k8s-res-023", "subPath 挂载子路径", "只挂载 ConfigMap 的单个键为文件", "volume", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: config\n      mountPath: /etc/app.yml\n      subPath: app.yml\n  volumes:\n  - name: config\n    configMap:\n      name: app-config",
			[]seedPoint{
				pt("mount", "/etc/app.yml", "mount_path", "easy", "挂载成的文件路径"),
				pt("sub", "subPath: app.yml", "volume", "medium", "只挂某键子路径"),
			},
			"`/etc/app.yml` / `subPath: app.yml`", "挂载单个文件", "`subPath` 只挂卷中的某个键", "`mountPath: /etc/app.yml; subPath: app.yml`",
			"subPath 让你只把卷里的一个文件（键）挂到目标路径，不影响更新。"),

		cfg("k8s-res-024", "securityContext 非 root", "以 UID 1000 非 root 运行容器", "security", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: secure\nspec:\n  securityContext:\n    runAsUser: 1000\n    runAsNonRoot: true\n  containers:\n  - name: app\n    image: nginx",
			[]seedPoint{
				pt("uid", "1000", "security", "easy", "运行用户 UID"),
				pt("nonroot", "runAsNonRoot: true", "security", "easy", "强制非 root 运行"),
			},
			"`1000` / `runAsNonRoot: true`", "以非 root 运行", "`securityContext` 设置运行身份", "`runAsUser: 1000` / `runAsNonRoot: true`",
			"以非 root 用户运行提升安全性；runAsNonRoot 校验镜像不以 root 运行。"),

		cfg("k8s-res-025", "只读根文件系统", "让容器根文件系统只读", "security", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: locked\nspec:\n  containers:\n  - name: app\n    image: nginx\n    securityContext:\n      readOnlyRootFilesystem: true",
			[]seedPoint{
				pt("ro", "readOnlyRootFilesystem: true", "security", "medium", "只读根文件系统选项"),
			},
			"`readOnlyRootFilesystem: true`", "根文件系统只读", "容器只读降低被篡改风险", "`readOnlyRootFilesystem: true`",
			"只读根文件系统防止写入持久化恶意内容，需配合可写卷使用。"),

		cfg("k8s-res-026", "多容器共享卷", "sidecar 容器与主容器共享 emptyDir", "multi_container", "hard",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: sidecar\nspec:\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: shared\n      mountPath: /var/log\n  - name: logger\n    image: fluentbit\n    volumeMounts:\n    - name: shared\n      mountPath: /logs\n  volumes:\n  - name: shared\n    emptyDir: {}",
			[]seedPoint{
				pt("mv", "/logs", "mount_path", "easy", "sidecar 的挂载路径"),
			},
			"`/logs`", "sidecar 共享卷", "同一卷被两个容器挂载即共享", "`mountPath: /logs`",
			"多容器共享同一卷实现日志/数据互通，是 sidecar 模式的基石。"),

		cfg("k8s-res-027", "initContainer 初始化", "用 init 容器准备数据后再启动主容器", "init_container", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: init-example\nspec:\n  initContainers:\n  - name: setup\n    image: busybox\n    command: [\"sh\", \"-c\", \"echo init > /work/data.txt\"]\n    volumeMounts:\n    - name: work\n      mountPath: /work\n  containers:\n  - name: app\n    image: nginx\n    volumeMounts:\n    - name: work\n      mountPath: /data\n  volumes:\n  - name: work\n    emptyDir: {}",
			[]seedPoint{
				pt("initname", "setup", "init_container", "easy", "init 容器名称"),
			},
			"`setup`", "init 容器", "`initContainers` 在主容器前依序运行", "`initContainers: - name: setup`",
			"init 容器完成前置初始化，主容器成功前不会启动。"),

		cfg("k8s-res-028", "Pod 终止宽限期", "设置优雅停机时间 30 秒", "lifecycle", "easy",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  terminationGracePeriodSeconds: 30\n  containers:\n  - name: app\n    image: nginx",
			[]seedPoint{
				pt("grace", "30", "lifecycle", "easy", "终止宽限期秒数"),
			},
			"`30`", "优雅停机时间", "`terminationGracePeriodSeconds` 给应用收尾时间", "`terminationGracePeriodSeconds: 30`",
			"删除 Pod 时给容器这段宽限期处理退出前清理，超时才强杀。"),

		cfg("k8s-res-029", "生命周期钩子 preStop", "删除前执行优雅退出命令", "lifecycle", "hard",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n  - name: app\n    image: nginx\n    lifecycle:\n      preStop:\n        exec:\n          command: [\"sh\", \"-c\", \"sleep 5; exit 0\"]",
			[]seedPoint{
				pt("hook", "preStop", "lifecycle", "hard", "删除前执行的钩子"),
			},
			"`preStop`", "生命周期钩子", "`preStop` 在终止宽限期内先执行", "`lifecycle: {preStop: {exec: ...}}`",
			"preStop 钩子在 SIGTERM 前执行，可做优雅下线/摘流量。"),

		cfg("k8s-res-030", "自定义 DNS 策略", "使用 ClusterFirstWithHostNet 的 DNS 策略", "dns", "medium",
			"apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  dnsPolicy: ClusterFirstWithHostNet\n  containers:\n  - name: app\n    image: nginx",
			[]seedPoint{
				pt("dns", "ClusterFirstWithHostNet", "dns", "medium", "DNS 策略值"),
			},
			"`ClusterFirstWithHostNet`", "DNS 策略", "配合 hostNetwork 使用集群 DNS", "`dnsPolicy: ClusterFirstWithHostNet`",
			"hostNetwork 时默认 DNS 失效，需用 ClusterFirstWithHostNet 保留集群解析。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, s := range seeds {
		ex, err := buildExercise(s)
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
