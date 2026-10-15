# Changelog

本项目的所有重要变更记录于此。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

## [Unreleased]

### Added
- CKA 认证题库（150 题，新增 `cka` 技术栈，5 个子库各 30 题，按 CNCF CKA 考纲五大域组织，命令/配置补全 + 答案匹配判题）：集群架构（kubeadm init·join·HA/版本升级全流程/etcd 快照备份·恢复·etcdutl/证书 check-expiration·renew/RBAC Role·RoleBinding·ClusterRole/CSR 审批/kubeconfig 上下文与默认命名空间/静态 Pod 目录/CNI 安装）、工作负载（Deployment 创建·滚动更新·回滚·扩缩/HPA v2/ConfigMap·Secret 三种注入方式/liveness·readiness 探针/资源 requests·limits/init 容器/sidecar 共享卷/DaemonSet/StatefulSet·volumeClaimTemplates/Job·CronJob/nodeSelector·亲和性/污点与容忍）、服务网络（Service 四类型/headless·ExternalName/手动 Endpoints/Ingress 多路径·TLS/CoreDNS Corefile/集群 FQDN/NetworkPolicy 默认拒绝·namespaceSelector·egress 含 DNS 放行/sessionAffinity/externalTrafficPolicy/命名端口/节点宿主网络 ip addr·route）、存储（PV hostPath·NFS·local·Block/PVC volumeName·selector 精确绑定/StorageClass 动态供给·默认类注解·WaitForFirstConsumer/PVC 在线扩容 patch/subPath 单文件挂载/accessModes·回收策略选择）、故障排查（NotReady 节点体检/kubelet systemctl·journalctl/drain·uncordon/crictl ps/CrashLoopBackOff --previous/事件排序/镜像 tag·selector 错配·探针 404·OOMKilled·PVC 名错·污点 Pending 修复/控制平面静态 Pod 修复/openssl 证书日期/DNS·Service 连通性验证/can-i --list 权限排查）
- 用户体系：注册 / 登录（PBKDF2-HMAC-SHA256 密码哈希 + 30 天会话 token）、退出登录、修改昵称 / 密码；进度、收藏、薄弱点、连续打卡按用户隔离
- 我的题库：用户勾选要练的 `tech::subcategory` 范围并持久化，首页与侧边栏按所选范围过滤（不选 = 全部）
- 题库动态加载：教练端 `POST /api/v1/admin/banks` 上传题库 JSON 即时生效；`data/bank` 目录热扫描（`-bank-watch`，默认 15s）自动加载新增 / 修改的题库文件
- 教练端用户管理：用户列表（作答 / 完成 / 连续天数）、创建、禁用 / 启用（禁用后无法登录）、重置密码（旧会话失效）、删除（级联清理进度 / 收藏 / 作答记录）
- 测试框架题库（180 题，6 个子库各 30 题，代码填空/命令/配置补全 + 答案匹配判题）：pytest（断言·raises·approx/fixture·yield·scope·params/conftest/parametrize·indirect/tmp_path/monkeypatch/capsys/mock/skip·xfail/pytest.ini·pyproject·CLI）、unittest（TestCase·断言族/setUp·tearDown/subTest/skip·expectedFailure/TestSuite·TestLoader·TextTestRunner/mock·patch·autospec/side_effect/assert_called_once_with·assert_has_calls/addCleanup/CLI）、selenium（Chrome 启动/By 定位/元素操作/implicitly_wait·WebDriverWait·EC/ActionChains/iframe·多窗口·alert/execute_script/headless/Remote Grid/cookies/Selenium Manager·Grid·docker-selenium 命令）、playwright（TS：getByRole·TestId·Text·Label/expect 断言族/route·fulfill·abort/storageState/beforeEach·test.step/soft 断言/request API/frameLocator/下载/playwright.config.ts/npx CLI）、cypress（JS：cy.visit·get·contains·within/type·click·check·select/should 断言族/intercept·as·wait/fixture·request·session/wrap·invoke·then/clock·tick/origin/Cypress.Commands.add/cypress.config.js/npx CLI）、locust（HttpUser·@task/wait_time·constant_pacing/on_start/client get·post/name·catch_response/events 监听/SequentialTaskSet·tag/FastHttpUser/LoadTestShape·tick/StopUser/headless·master·worker·csv·html CLI/locust.conf）
- OpenStack 题库（200 题，10 个子库各 20 题，命令/配置补全 + 答案匹配判题）：Nova（实例生命周期/迁移/配额/nova.conf）、Neutron（网络·子网·路由/安全组/Floating IP/ML2·OVS·L3 配置）、Cinder（卷·快照·备份/类型与 QoS/LVM·Ceph 后端）、Swift（上传下载/大对象/临时 URL/ring/中间件 pipeline）、TripleO（undercloud/overcloud 部署/节点自省/roles_data/升级更新）、Keystone（token/项目·用户·角色/catalog/endpoint/fernet/policy）、Glance（镜像导入导出/共享/格式/多后端/缓存）、运维（服务巡检/日志/DB·MQ/高可用/故障疏散/备份）、升级（版本流/三阶段升级/DB 迁移/镜像 tag/回滚）、Ceph（集群状态/pool·RBD/CRUSH/认证/与 Cinder·Glance·Nova 对接）
- Python 进阶题库（90 题，6 个子分类各 15 题，全部真实执行判题）：装饰器（wraps/带参装饰器/叠加顺序/类装饰器/lru_cache/singledispatch/property/重试）、迭代器与生成器（iter·next/自定义可迭代与迭代器/yield/yield from/send/throw/close/islice/groupby）、描述符（__get__·__set__·__delete__/__set_name__/数据与非数据优先级/惰性缓存/classmethod 原理）、上下文管理器（__enter__·__exit__/异常抑制/contextmanager/suppress/closing/redirect_stdout/ExitStack）、元类与类机制（type() 动态建类/元类 __new__·__init__/__init_subclass__/MRO·super/__slots__/ABCMeta/__instancecheck__）、魔术方法（__repr__·__str__/__eq__·__hash__/__len__·__bool__/__getitem__·__contains__/__call__/__getattr__·__getattribute__/运算符重载/total_ordering/__missing__）
- 数据结构系列题库（60 题，Python，6 个子分类各 10 题）：数组（双指针/二分/前缀和/差分/滑动窗口）、链表（反转/快慢指针/合并/环检测/两数相加）、栈与队列（括号匹配/逆波兰/最小栈/循环队列/单调栈与单调队列）、哈希表（手写哈希/两数之和/异位词分组/开放寻址/LRU 缓存）、树（遍历/BST 插入查找/验证/翻转/路径总和/最小堆/Trie）、图（邻接表/BFS/DFS/连通分量/最短路/拓扑排序/环检测/并查集/岛屿数量/Dijkstra），全部真实执行判题
- Go 类型系统题库（30 题，Go 子分类「类型系统」）：零值与类型推断/iota/定义类型与别名/显式转换/rune 与字节/类型断言与 type switch/结构体方法集/嵌入与标签/接口隐式实现/nil 接口陷阱/any/error/接口组合/泛型约束/指针与引用语义，真实 `go run` 执行判题
- Agentic AI 开发题库（50 题，Python 子分类「Agent 开发」）：消息与采样参数、工具 Schema/注册分派/异常兜底、Agent 循环与 ReAct 轨迹、滑动窗口/摘要/向量检索/RAG、结构化输出校验与重试、多智能体路由与 handoff、退避重试/成本核算/日志，全部真实执行判题
- 训练点 `order_group` 顺序可互换分组：同组判题做双射匹配，行序无关的配置行（如 HCL 块内属性）写反同样判对；重复答案仍只能匹配一个训练点
- SQL 基础题库（50 题）：SELECT/WHERE/排序分页/聚合分组/JOIN/子查询/DML/DDL，统一示例库出题
- Python 协程专项题库（20 题）：async/await/create_task/gather/超时异常/异步迭代器/Queue/Semaphore
- Python 多线程专项题库（20 题）：Thread/join/Lock·RLock/Semaphore/Event/Barrier/threading.local/queue.Queue/Timer/线程池
- BigQuery 题库（61 题）拆为两个子库：「SQL 基础」（35 题通用语法）与「BigQuery 特有」（26 题：数组/STRUCT/UNNEST、JSON、窗口·QUALIFY、分区聚类、通配表、INFORMATION_SCHEMA、脚本变量、采样、时间旅行）
- DevOps 题库（140 题）拆为四个子库：「terraform」（30）、「ansible」（30）及并入的「docker」（50）、「prometheus」（30）
- AWS 题库（180 题）六个子库各 30 题：ec2/iam/eks/s3/vpc/rds
- 教练端接口令牌鉴权：`-admin-token` 或 `CODEGYM_ADMIN_TOKEN` 启用后所有 `/api/v1/admin/*` 需携带令牌
- Docker 部署：`Dockerfile.backend` / `Dockerfile.frontend` / `docker-compose.yml`
- GitHub Actions CI：后端 -race 测试、前端测试与构建、题库审计

### Changed
- 用户身份解析由 `X-User-Id` 请求头改为会话 token（`Authorization: Bearer`），未登录请求回退到共享的 demo 用户；e2e 测试同步改为注册 / 登录后携带 token
- OpenStack 题库训练点全部细化为「小空」：原先一个空要输入整条命令（最长 11 个词），现拆成子命令/参数/配置键值等单元（每个空 ≤2 个词，1300 个训练点），各难度抽空数恢复为入门 1-2、简单 2-3、中等 3-4、困难 4-6、专家 6-10
- Python 进阶题库 90 题全部扩充为每题 6 个训练点（2 入门 + 1 简单 + 1 中等 + 1 困难 + 1 专家），入门到专家各难度档都有足够空位可练
- Go 并发专项（50 题）拆分为 3 个子库：通信机制（25：goroutine/WaitGroup/channel/select/缓冲死锁）、同步与原子（13：Mutex·RWMutex/Once/atomic/Cond/Pool/数据竞争与并发 map 修复）、并发模式（12：生成器/Worker Pool/Fan-in·out/Pipeline/生产者消费者/context/Ticker/优雅关闭），题目内容不变
- Bug 修复题型训练点提示改用题目级四级辅助（原按类别生成的提示与填空内容错位）
- channel 类训练点 L2 提示补充 comma-ok 说明；新增 map_write/critical_section 等类别映射

### Fixed
- 执行失败路径 `fill_results` 返回 null 导致前端评分卡崩溃（现恒为空数组）
- 前端填空答案提取被后续代码段文字截断的回归问题
- 编辑器填空区间追踪三连修：删除答案后重新输入不被识别、退格多删破坏代码结构后答案错位、单填空整段粘贴吞并固定代码
- 编辑器超长行被裁剪：按内容撑开宽度出横向滚动，行号栏 sticky 钉在可视区左缘；同时修复超过 560px 的长代码纵向滚动后高亮与行号消失

## [0.2.0] - 2026-08

### Added
- 题库全面 JSON 化到 `backend/data/bank/`，不改代码即可增删改题目
- Kubernetes 调度专项（30 题）、OpenShift 对比专项（30 题）、Prometheus 基础（30 题）
- Go 并发专项（50 题，真实 go run 判题）+ 自动补训练点机制
- 教练端 AI 出题（Ollama）+ Ollama 设置页面化 + 执行环境管理
- 侧栏收藏视图、按代码位置排列评分表

## [0.1.0] - 2026-08

### Added
- CodeGym MVP：填空-运行-验证闭环、负重分级、四级辅助、评分评级
- Python / Go / JavaScript / Docker / OpenShift / Kubernetes 首批题库
- 训练进度、薄弱部位、收藏、导入导出、教练端 CRUD
