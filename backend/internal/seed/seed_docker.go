package seed

import (
	"codegym/backend/internal/model"
)

// toDockerSeed converts a Docker seed (same shape as k8s) to a seedExercise
// under the "docker" tech stack and the "基础" subcategory.
func toDockerSeed(k k8sBasicSeed) seedExercise {
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
		techStack: "docker", weight: k.weight, category: k.category, subcategory: "基础",
		fullCode: k.code, expected: k.expected, points: points,
		hints: goHint(k.hintL1, k.hintL2, k.hintL3, "正确答案："+k.code), explanation: k.explanation,
	}
}

// DockerBasics returns the 50-question Docker basic bank (命令 + Dockerfile + compose).
func DockerBasics() ([]model.Exercise, error) {
	pt := func(id, text, category, weight, hint string) seedPoint {
		return seedPoint{id: id, text: text, category: category, minWeight: weight,
			weightScore: 1, hint: hint, fillType: "text"}
	}
	cmd := func(id, title, desc, category, code string, p seedPoint, l1, l2, l3, expl string) k8sBasicSeed {
		return k8sBasicSeed{id: id, title: title, desc: desc, kind: "command", category: category,
			weight: "easy", code: code, points: []seedPoint{p}, hintL1: l1, hintL2: l2, hintL3: l3,
			explanation: expl}
	}
	df := func(id, title, desc, category, code string, points []seedPoint, l1, l2, l3, _l4, expl string) k8sBasicSeed {
		return k8sBasicSeed{id: id, title: title, desc: desc, kind: "config", category: category,
			weight: "medium", code: code, points: points, hintL1: l1, hintL2: l2, hintL3: l3,
			explanation: expl}
	}

	seeds := []k8sBasicSeed{
		// ---- docker 命令 ----
		cmd("docker-basics-001", "构建镜像", "从当前目录构建镜像并打标签", "docker_build",
			"docker build -t myapp:v1 .",
			pt("flag", "-t myapp:v1", "command_flag", "beginner", "给镜像打标签"),
			"构建镜像的命令", "`-t` 指定镜像名:标签", "`docker build -t myapp:v1 .`",
			"`docker build` 按 Dockerfile 构建镜像，`-t` 打标签，`.` 指明构建上下文。"),

		cmd("docker-basics-002", "后台运行容器", "以后台守护方式运行 nginx", "docker_run",
			"docker run -d nginx",
			pt("flag", "-d", "command_flag", "beginner", "后台运行选项"),
			"后台运行的选项", "`-d` 后台守护运行", "`docker run -d nginx`",
			"`-d` 让容器在后台运行，前台用 `-it`（交互式）或直接运行。"),

		cmd("docker-basics-003", "交互运行", "以交互式终端运行 Ubuntu", "docker_run",
			"docker run -it ubuntu bash",
			pt("flag", "-it", "command_flag", "easy", "交互式终端选项"),
			"交互式运行的选项", "`-it` = `-i -t`（stdin + tty）", "`docker run -it ubuntu bash`",
			"`-it` 连接标准输入并分配伪终端，用于进入容器内 shell。"),

		cmd("docker-basics-004", "端口映射", "把容器 80 映射到主机 8080", "docker_run",
			"docker run -p 8080:80 nginx",
			pt("flag", "-p 8080:80", "command_flag", "easy", "端口映射"),
			"端口映射的选项", "`-p 主机端口:容器端口`", "`docker run -p 8080:80 nginx`",
			"`-p` 发布端口到宿主机，格式 `主机端口:容器端口`。"),

		cmd("docker-basics-005", "挂载数据卷", "把宿主机目录挂载进容器", "docker_run",
			"docker run -v /data:/data nginx",
			pt("flag", "-v /data:/data", "command_flag", "medium", "卷挂载"),
			"挂载卷的选项", "`-v 宿主机:容器内`", "`docker run -v /data:/data nginx`",
			"`-v` 把宿主机目录/卷挂载到容器，用于持久化数据。"),

		cmd("docker-basics-006", "设置环境变量", "运行容器并设置环境变量", "docker_run",
			"docker run -e MODE=prod nginx",
			pt("flag", "-e MODE=prod", "command_flag", "easy", "环境变量"),
			"设置环境变量的选项", "`-e 键=值`", "`docker run -e MODE=prod nginx`",
			"`-e` 向容器注入环境变量；也可从文件 `--env-file` 读取。"),

		cmd("docker-basics-007", "指定容器名", "运行容器并命名为 web", "docker_run",
			"docker run --name web nginx",
			pt("flag", "--name web", "command_flag", "easy", "容器名"),
			"指定容器名的选项", "`--name <名>`", "`docker run --name web nginx`",
			"`--name` 给容器命名，便于用名字管理（stop/rm 等）。"),

		cmd("docker-basics-008", "拉取镜像", "拉取 nginx:1.25 镜像", "docker_pull",
			"docker pull nginx:1.25",
			pt("verb", "pull", "command_verb", "beginner", "拉取镜像命令"),
			"拉取镜像的命令", "`docker pull <镜像>`", "`docker pull nginx:1.25`",
			"`docker pull` 从仓库拉取镜像到本地。"),

		cmd("docker-basics-009", "查看镜像", "列出本地镜像", "docker_ps",
			"docker images",
			pt("verb", "images", "command_verb", "beginner", "查看镜像命令"),
			"查看镜像的命令", "`docker images`", "`docker images`",
			"`docker images` 列出本地已拉取/构建的镜像。"),

		cmd("docker-basics-010", "查看运行容器", "列出正在运行的容器", "docker_ps",
			"docker ps",
			pt("verb", "ps", "command_verb", "beginner", "查看进程/容器"),
			"查看运行容器的命令", "`docker ps`", "`docker ps`",
			"`docker ps` 默认只看运行中的容器。"),

		cmd("docker-basics-011", "查看所有容器", "列出所有容器（含已停止）", "docker_ps",
			"docker ps -a",
			pt("flag", "-a", "command_flag", "easy", "显示所有容器"),
			"显示所有容器的选项", "`-a` 包含停止的容器", "`docker ps -a`",
			"`docker ps -a` 包含已停止的容器，前一位 `--all`。"),

		cmd("docker-basics-012", "停止容器", "停止 web 容器", "docker_stop",
			"docker stop web",
			pt("verb", "stop", "command_verb", "beginner", "停止命令"),
			"停止容器的命令", "`docker stop <名>`", "`docker stop web`",
			"`docker stop` 发送 SIGTERM 优雅停止容器，`kill` 则是 SIGKILL。"),

		cmd("docker-basics-013", "启动容器", "启动已停止的 web 容器", "docker_start",
			"docker start web",
			pt("verb", "start", "command_verb", "beginner", "启动命令"),
			"启动容器的命令", "`docker start <名>`", "`docker start web`",
			"`docker start` 启动已存在的（已停止）容器，不新建。"),

		cmd("docker-basics-014", "删除容器", "删除 web 容器", "docker_rm",
			"docker rm web",
			pt("verb", "rm", "command_verb", "beginner", "删除容器命令"),
			"删除容器的命令", "`docker rm <名>`", "`docker rm web`",
			"`docker rm` 删除已停止的容器；运行中的需 `-f` 强制。"),

		cmd("docker-basics-015", "强制删除运行容器", "强制删除正在运行的容器", "docker_rm",
			"docker rm -f web",
			pt("flag", "-f", "command_flag", "easy", "强制删除"),
			"强制删除的选项", "`-f` 强制删除运行中的", "`docker rm -f web`",
			"`docker rm -f` 对运行中的容器先停止再删除。"),

		cmd("docker-basics-016", "删除镜像", "删除本地 nginx 镜像", "docker_rmi",
			"docker rmi nginx",
			pt("verb", "rmi", "command_verb", "easy", "删除镜像命令"),
			"删除镜像的命令", "`docker rmi <镜像>`", "`docker rmi nginx`",
			"`docker rmi` 删除本地镜像；被使用中的镜像无法删除。"),

		cmd("docker-basics-017", "进入容器", "进入 web 容器的 bash", "docker_exec",
			"docker exec -it web bash",
			pt("verb", "exec", "command_verb", "easy", "在容器内执行"),
			"进入容器执行的命令", "`docker exec -it <名> bash`", "`docker exec -it web bash`",
			"`docker exec` 在运行中的容器内执行命令（不同于 `run` 新建容器）。"),

		cmd("docker-basics-018", "查看容器日志", "查看 web 容器日志", "docker_logs",
			"docker logs web",
			pt("verb", "logs", "command_verb", "beginner", "查看日志"),
			"查看日志的命令", "`docker logs <名>`", "`docker logs web`",
			"`docker logs` 查看容器 stdout/stderr 日志。"),

		cmd("docker-basics-019", "实时跟随日志", "实时输出 web 容器日志", "docker_logs",
			"docker logs -f web",
			pt("flag", "-f", "command_flag", "easy", "跟随日志"),
			"实时日志的选项", "`-f` 实时跟随", "`docker logs -f web`",
			"`docker logs -f` 流式跟随日志，类似 `tail -f`。"),

		cmd("docker-basics-020", "容器详情", "查看 web 容器详细配置", "docker_inspect",
			"docker inspect web",
			pt("verb", "inspect", "command_verb", "medium", "查看详情"),
			"查看容器详情的命令", "`docker inspect <名>`", "`docker inspect web`",
			"`docker inspect` 输出容器/镜像的底层 JSON 元数据。"),

		cmd("docker-basics-021", "查看端口映射", "查看 web 容器的端口映射", "docker_port",
			"docker port web",
			pt("verb", "port", "command_verb", "medium", "查看端口"),
			"查看端口映射的命令", "`docker port <名>`", "`docker port web`",
			"`docker port` 显示容器端口与宿主机端口的映射关系。"),

		cmd("docker-basics-022", "复制文件", "把本地文件复制进容器", "docker_cp",
			"docker cp app.txt web:/tmp",
			pt("verb", "cp", "command_verb", "medium", "复制到容器"),
			"复制文件的命令", "`docker cp 本地 <名>:路径`", "`docker cp app.txt web:/tmp`",
			"`docker cp` 在宿主机和容器之间复制文件/目录。"),

		cmd("docker-basics-023", "资源统计", "实时查看容器资源占用", "docker_stats",
			"docker stats",
			pt("verb", "stats", "command_verb", "medium", "资源统计"),
			"查看资源占用的命令", "`docker stats`", "`docker stats`",
			"`docker stats` 实时显示容器的 CPU/内存/网络占用。"),

		cmd("docker-basics-024", "查看网络", "列出 Docker 网络", "docker_network",
			"docker network ls",
			pt("verb", "network ls", "command_verb", "medium", "网络子命令"),
			"查看网络的命令", "`docker network ls`", "`docker network ls`",
			"`docker network` 管理容器网络（bridge/host/none 等）。"),

		cmd("docker-basics-025", "查看数据卷", "列出 Docker 数据卷", "docker_volume",
			"docker volume ls",
			pt("verb", "volume ls", "command_verb", "medium", "卷子命令"),
			"查看卷的命令", "`docker volume ls`", "`docker volume ls`",
			"`docker volume` 管理命名卷，用于持久化数据。"),

		cmd("docker-basics-026", "打镜像标签", "给镜像打新标签", "docker_tag",
			"docker tag myapp myrepo/myapp:v1",
			pt("verb", "tag", "command_verb", "medium", "打标签"),
			"打标签的命令", "`docker tag 原 新`", "`docker tag myapp myrepo/myapp:v1`",
			"`docker tag` 为镜像添加别名/标签，常用于推送前重命名。"),

		cmd("docker-basics-027", "推送镜像", "把镜像推送到仓库", "docker_push",
			"docker push myrepo/myapp:v1",
			pt("verb", "push", "command_verb", "medium", "推送镜像"),
			"推送镜像的命令", "`docker push <仓库>/<名>:<tag>`", "`docker push myrepo/myapp:v1`",
			"`docker push` 把本地镜像推送到远程镜像仓库。"),

		cmd("docker-basics-028", "保存镜像", "把镜像保存为 tar 文件", "docker_save",
			"docker save -o app.tar myapp:v1",
			pt("flag", "-o app.tar", "command_flag", "hard", "输出文件"),
			"保存镜像的命令", "`docker save -o <文件> <镜像>`", "`docker save -o app.tar myapp:v1`",
			"`docker save` 把镜像导出为 tar 便于离线迁移；`load` 反向导入。"),

		cmd("docker-basics-029", "导出容器文件系统", "把容器导出为 tar", "docker_export",
			"docker export web > web.tar",
			pt("verb", "export", "command_verb", "hard", "导出容器"),
			"导出容器的命令", "`docker export <名>`", "`docker export web`",
			"`export` 导出容器文件系统（不含元数据），与 `save`（镜像）不同。"),

		cmd("docker-basics-030", "清理未用资源", "清理未使用的镜像/容器/网络", "docker_prune",
			"docker system prune",
			pt("verb", "system prune", "command_verb", "medium", "系统清理"),
			"清理资源的命令", "`docker system prune`", "`docker system prune`",
			"`system prune` 清理未使用的对象；`-a` 追加清理悬空镜像。"),

		cmd("docker-basics-031", "compose 启动", "后台启动 compose 服务", "docker_compose",
			"docker compose up -d",
			pt("verb", "up", "command_verb", "easy", "启动服务"),
			"compose 启动命令", "`docker compose up -d`", "`docker compose up -d`",
			"`docker compose up` 按 docker-compose.yml 启动服务，`-d` 后台运行。"),

		cmd("docker-basics-032", "compose 停止", "停止并移除 compose 服务", "docker_compose",
			"docker compose down",
			pt("verb", "down", "command_verb", "easy", "停止服务"),
			"compose 停止命令", "`docker compose down`", "`docker compose down`",
			"`docker compose down` 停止并移除由 compose 创建的服务与网络。"),

		// ---- Dockerfile ----
		df("docker-basics-033", "Dockerfile 基础镜像", "用 FROM 指定基础镜像", "dockerfile",
			"FROM python:3.11-slim",
			[]seedPoint{
				pt("base", "python:3.11-slim", "base_image", "beginner", "基础镜像名及标签"),
			},
			"`FROM python:3.11-slim`", "Dockerfile 第一行", "`FROM` 指定基础镜像", "`FROM python:3.11-slim`",
			"`FROM` 是 Dockerfile 第一个指令，基于某个镜像构建。"),

		df("docker-basics-034", "Dockerfile 执行命令", "构建时安装依赖", "dockerfile",
			"FROM ubuntu:22.04\nRUN apt-get update && apt-get install -y curl",
			[]seedPoint{
				pt("run", "apt-get update", "run_cmd", "easy", "构建期执行的命令"),
				pt("install", "apt-get install -y curl", "run_cmd", "medium", "安装软件包"),
			},
			"`apt-get update` / `apt-get install -y curl`", "RUN 指令", "`RUN` 在构建镜像时执行命令", "`RUN apt-get update && ...`",
			"`RUN` 在构建阶段执行命令，`&&` 串联可减少镜像层。"),

		df("docker-basics-035", "COPY 拷贝文件", "把本地目录拷贝进镜像", "dockerfile",
			"FROM python:3.11-slim\nWORKDIR /app\nCOPY . .",
			[]seedPoint{
				pt("copy", "COPY . .", "copy", "easy", "拷贝指令"),
				pt("workdir", "/app", "workdir", "medium", "工作目录"),
			},
			"`COPY . .` / `/app`", "COPY 与 WORKDIR", "`COPY` 把构建上下文文件拷入镜像", "`COPY . .`",
			"`WORKDIR` 设定工作目录，`COPY 源 目标` 拷贝文件进镜像。"),

		df("docker-basics-036", "ADD 拷贝与解压", "用 ADD 拷贝并可自动解压", "dockerfile",
			"FROM ubuntu:22.04\nADD app.tar.gz /app/",
			[]seedPoint{
				pt("add", "ADD app.tar.gz /app/", "add", "medium", "拷贝解压指令"),
			},
			"`ADD app.tar.gz /app/`", "ADD 指令", "`ADD` 支持 URL 与自动解压 tar", "`ADD app.tar.gz /app/`",
			"`ADD` 类似 COPY，额外支持 URL 与 tar 解压；纯拷贝优先用 COPY。"),

		df("docker-basics-037", "CMD 默认命令", "设容器默认启动命令", "dockerfile",
			"FROM nginx:1.25\nCMD [\"nginx\", \"-g\", \"daemon off;\"]",
			[]seedPoint{
				pt("cmd", "CMD", "cmd", "easy", "默认命令指令"),
				pt("arr", "[\"nginx\", \"-g\", \"daemon off;\"]", "cmd", "medium", "exec 形式命令"),
			},
			"`CMD` / `[\"nginx\", \"-g\", \"daemon off;\"]`", "CMD 指令", "`CMD` 定义默认启动命令，可被 `run` 覆盖", "`CMD [\"nginx\", ...]`",
			"`CMD` 提供容器默认启动命令，推荐 exec 形式（方括号数组）。"),

		df("docker-basics-038", "ENTRYPOINT 入口", "设置不可覆盖的入口命令", "dockerfile",
			"FROM python:3.11-slim\nENTRYPOINT [\"python\", \"app.py\"]",
			[]seedPoint{
				pt("entry", "ENTRYPOINT [\"python\", \"app.py\"]", "entrypoint", "medium", "入口命令"),
			},
			"`ENTRYPOINT [\"python\", \"app.py\"]`", "ENTRYPOINT 指令", "`ENTRYPOINT` 是入口命令，`CMD` 可作参数补充", "`ENTRYPOINT [\"python\", \"app.py\"]`",
			"ENTRYPOINT 固定容器入口命令，不会因 `run` 传参而被覆盖。"),

		df("docker-basics-039", "EXPOSE 端口", "声明容器监听端口", "dockerfile",
			"FROM nginx:1.25\nEXPOSE 80",
			[]seedPoint{
				pt("port", "80", "expose_port", "easy", "监听的容器端口"),
			},
			"`80`", "EXPOSE 指令", "`EXPOSE` 声明容器监听的端口（仅说明）", "`EXPOSE 80`",
			"`EXPOSE` 仅是文档性质声明，真正发布端口用 `docker run -p`。"),

		df("docker-basics-040", "ENV 环境变量", "在镜像内设置环境变量", "dockerfile",
			"FROM python:3.11-slim\nENV PYTHONUNBUFFERED=1",
			[]seedPoint{
				pt("env", "PYTHONUNBUFFERED=1", "env", "easy", "环境变量键值"),
			},
			"`PYTHONUNBUFFERED=1`", "ENV 指令", "`ENV` 设置环境变量并固化进镜像", "`ENV PYTHONUNBUFFERED=1`",
			"`ENV` 定义环境变量，运行与构建阶段均可用。"),

		df("docker-basics-041", "ARG 构建参数", "声明可传参的构建变量", "dockerfile",
			"FROM python:3.11-slim\nARG VERSION=1.0\nRUN echo $VERSION",
			[]seedPoint{
				pt("arg", "VERSION=1.0", "arg", "hard", "构建参数默认值"),
			},
			"`ARG VERSION=1.0`", "ARG 指令", "`ARG` 声明构建期参数，`--build-arg` 可传入", "`ARG VERSION=1.0`",
			"`ARG` 只在构建期有效（不进入镜像），用 `--build-arg` 覆盖。"),

		df("docker-basics-042", "USER 运行用户", "以非 root 用户运行", "dockerfile",
			"FROM ubuntu:22.04\nRUN groupadd -r app && useradd -r -g app app\nUSER app",
			[]seedPoint{
				pt("user", "app", "user", "easy", "运行用户"),
			},
			"`USER app`", "USER 指令", "`USER` 切换后续 RUN/CMD 的运行用户", "`USER app`",
			"`USER` 设置镜像内后续操作以非 root 用户执行，提升安全性。"),

		df("docker-basics-043", "VOLUME 声明卷", "声明匿名数据卷", "dockerfile",
			"FROM nginx:1.25\nVOLUME /var/www/html",
			[]seedPoint{
				pt("volume", "/var/www/html", "volume", "medium", "声明的卷路径"),
			},
			"`/var/www/html`", "VOLUME 指令", "`VOLUME` 声明匿名卷目录", "`VOLUME /var/www/html`",
			"`VOLUME` 声明挂载点，容器运行时会创建匿名卷存储该目录数据。"),

		df("docker-basics-044", "HEALTHCHECK 健康检查", "给镜像加健康检查", "dockerfile",
			"FROM nginx:1.25\nHEALTHCHECK CMD curl -f http://localhost/ || exit 1",
			[]seedPoint{
				pt("health", "curl -f http://localhost/", "healthcheck", "hard", "检查命令"),
			},
			"`curl -f http://localhost/`", "HEALTHCHECK 指令", "`HEALTHCHECK CMD <命令>` 定义健康探测", "`HEALTHCHECK CMD curl ...`",
			"HEALTHCHECK 让 Docker 定时探测容器健康，失败会标记 unhealthy。"),

		df("docker-basics-045", "LABEL 元数据", "给镜像打标签元数据", "dockerfile",
			"FROM nginx:1.25\nLABEL maintainer=\"ops@example.com\"",
			[]seedPoint{
				pt("label", "maintainer=\"ops@example.com\"", "label", "hard", "标签键值"),
			},
			"`maintainer=\"ops@example.com\"`", "LABEL 指令", "`LABEL` 添加键值元数据", "`LABEL maintainer=\"...\"`",
			"`LABEL` 为镜像添加键值元数据（作者/版本等）。"),

		// ---- docker-compose.yml ----
		df("docker-basics-046", "compose 服务与镜像", "定义一个服务并指定镜像", "compose",
			"services:\n  web:\n    image: nginx:1.25",
			[]seedPoint{
				pt("img", "nginx:1.25", "image", "beginner", "服务所用的镜像"),
			},
			"`nginx:1.25`", "compose 服务定义", "`services` 下定义服务，`image` 指定镜像", "`image: nginx:1.25`",
			"compose 文件以 `services` 定义各服务，`image` 声明所用镜像。"),

		df("docker-basics-047", "compose 端口映射", "映射服务端口", "compose",
			"services:\n  web:\n    image: nginx:1.25\n    ports:\n    - \"8080:80\"",
			[]seedPoint{
				pt("ports", "\"8080:80\"", "ports", "easy", "主机端口:容器端口"),
			},
			"`\"8080:80\"`", "compose 端口映射", "`ports` 列表格式 `\"主机:容器\"`", "`ports: - \"8080:80\"`",
			"`ports` 用 `\"主机端口:容器端口\"` 发布服务端口。"),

		df("docker-basics-048", "compose 数据卷", "为服务挂载数据卷", "compose",
			"services:\n  db:\n    image: postgres:15\n    volumes:\n    - db-data:/var/lib/postgresql/data\nvolumes:\n  db-data:",
			[]seedPoint{
				pt("vol", "db-data:/var/lib/postgresql/data", "volumes", "medium", "卷:容器路径"),
				pt("decl", "db-data:", "volumes_decl", "hard", "命名卷声明"),
			},
			"`db-data:/var/lib/postgresql/data` / `db-data:`", "compose 数据卷", "`volumes` 挂载，底部声明命名卷", "`- db-data:/var/lib/postgresql/data`",
			"先挂载 `卷:容器路径`，再在顶层 `volumes` 声明命名卷。"),

		df("docker-basics-049", "compose 环境变量", "给服务设置环境变量", "compose",
			"services:\n  web:\n    image: myapp\n    environment:\n      MODE: prod\n      DEBUG: \"false\"",
			[]seedPoint{
				pt("mode", "prod", "env", "easy", "环境变量值"),
			},
			"`MODE: prod`", "compose 环境变量", "`environment` 键值列表", "`environment: {MODE: prod}`",
			"`environment` 以键值设置服务环境变量。"),

		df("docker-basics-050", "compose 重启与依赖", "设置重启策略与依赖顺序", "compose",
			"services:\n  app:\n    image: myapp\n    restart: unless-stopped\n    depends_on:\n    - db\n  db:\n    image: postgres:15",
			[]seedPoint{
				pt("restart", "unless-stopped", "restart", "medium", "重启策略"),
				pt("dep", "- db", "depends_on", "medium", "依赖的服务"),
			},
			"`unless-stopped` / `- db`", "restart 与 depends_on", "`restart` 决定重启策略，`depends_on` 声明依赖", "`restart: unless-stopped`",
			"`restart` 控制崩溃/重启行为，`depends_on` 控制启动顺序。"),
	}

	out := make([]model.Exercise, 0, len(seeds))
	for _, k := range seeds {
		ex, err := buildExercise(toDockerSeed(k))
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
