# 123pan 私人图床系统

基于 **123云盘开放平台 API** 开发的轻量、高性能私人图床。采用 Go (Gin) 后端 + 原生 HTML/JS/CSS 前端构建，剥离繁重的外部框架前端，方便内网、软路由或云服务器跨平台部署。

## ✨ 核心特性

- **简易美观的控制台**：内嵌现代化响应式前端，支持暗黑模式和玻璃拟态 UI。
- **极速上传体验**：支持多文件拖拽、点击选择，以及原生兼容的**全局 Ctrl+V 剪贴板一键粘贴直传**。
- **高并发与流式防御**：基于 123pan 复杂多步骤 API 深层对接；动态计算文件 MD5，抛弃无脑内存读取（防止服务内存溢出宕机）。
- **全自动鉴权保活**：无需定时折腾 Token。提供 ClientID 和 Secret 后，系统会自动管理 `access_token` 并做提前无感更新。
- **最高级安全验证**：全部 API 支持 `api_token` 自定义密码保护。阻断防侧信道劫持，杜绝接口裸奔被陌生人滥用。

---

## ⚙️ 快速开始

### 1. 准备条件 (⚠️ 必看：平台鉴权机制收费前置)
- 本地或服务器已部署好 **Go 编译环境**。
- **必须解锁 API 调度权限**：由于 123云盘架构政策重大调整，目前开放平台的直链及 API 提取额度已全线转为付费高级属性。您 **必须** 首先前往 123云盘主站的 [开发者权益专区](https://www.123pan.com/member?source_page=img_center) 购买订阅对应的“开发者权益包”。
- 购买并激活底层接口权益后，才能前往 [123云盘开放平台控制台](https://platform.123pan.com/) 创建应用并拿到真正具备流转效力的 `Client ID` 与 `Client Secret`。（**注**：若未打通付费权限通道直接填入尝试调用，后台请求将被严防死守并无限提示 `获取列表异常: 无效的登录信息`。）

### 2. 补齐配置
进入项目的 `conf` 文件夹，修改 `config.yaml`（请确保该文件已创建并粘贴如下必要字段）：

```yaml
# 服务器端口
port: 8080

# 123云盘开放平台应用配置 (必需)
client_id: "你的应用客户端ID"
client_secret: "你的应用客户端Secret"

# 123云盘专属存放目录的 ID (根目录可留空 "")
parent_file_id: "你的文件夹ID"

# 系统访问保护密码！如果挂至公网，强烈建议自行设置复杂的密码！
api_token: "替换成你设定的强密码"
```

### 3. 上线运行
在 `imagehost` 根目录控制台执行：
```bash
go mod tidy
go run cmd/main.go
```

### 4. 跨平台编译打包 (交叉编译指南)
得益于 Go 强大的跨平台支持以及本项目采用的 **完全内嵌静态打包 (`go:embed`)** 机制，你不再需要搬运 `static` 文件夹。将应用编译出体积极小、没有任何外部依赖的单一核心可执行文件（唯一需要同处一个目录的是配置文件 `conf/config.yaml` ）。

以下是不同目标平台的编译命令（如果报错，请确保没有在命令中附带多余符号）：

- **Windows 本地平台 (如您目前系统)**: 
  ```bash
  go build -o imagehost.exe ./cmd/main.go
  ```
- **Linux 服务器 / 软路由 (x86_64 常见架构)**:
  ```powershell
  $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o imagehost-linux ./cmd/main.go
  ```
  *(注：如果是在纯 Linux/macOS 终端编译则为 `GOOS=linux GOARCH=amd64 go build -o imagehost-linux ./cmd/main.go`，下同)*
- **Linux 盒子 / 树莓派 (ARM64 轻量级结构)**:
  ```powershell
  $env:GOOS="linux"; $env:GOARCH="arm64"; go build -o imagehost-linux-arm ./cmd/main.go
  ```
- **macOS (M1/M2/M3 Apple Silicon 原生架构)**:
  ```powershell
  $env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o imagehost-macos-arm ./cmd/main.go
  ```
- **macOS (老的 Intel 原生架构)**:
  ```powershell
  $env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o imagehost-macos ./cmd/main.go
  ```

启动成功后，浏览器中打卡 `http://localhost:8080/` 即可看到私有云图床台。若启用了 `api_token`，首页将要求输入验证。

---

## 🐳 Docker 云原生极简部署

如果你使用的是 NAS（如群晖、威联通）、软路由系统或原生携带 Docker 的云服务器，我们极力推荐使用此方案。

本项目完全遵循 **云原生十二要素 (12-Factor App)** 规范。全核心参数支持通过 **环境变量 (Environment Variables)** 纯净直连，彻底告别繁杂凌乱的配置文件管理！

#### 👉 使用 docker-compose 启动

找一个空目录，新建 `docker-compose.yml` 文件并粘贴以下内容：

```yaml
version: '3.8'

services:
  imagehost:
    # 替换为你刚才在 DockerHub 推送的独特镜像名！
    image: 你的用户名/123pan-imagehost:latest 
    container_name: 123pan-imagehost
    restart: unless-stopped
    ports:
      - "8080:8080" # 左侧是你要暴露在外网的访问端口
    environment:
      # 🌍 全量环保注入，零文件残留！
      - PORT=8080
      - CLIENT_ID=请去123pan开放平台申请必填
      - CLIENT_SECRET=请去123pan开放平台申请必填
      - PARENT_FILE_ID= # 专用上传目标的文件夹ID(如果传到根目录则删掉注释留空)
      - API_TOKEN=PRIVATE_123_KEY # 拦截门票：强烈建议设为复杂密码防御公网盗刷
      - TZ=Asia/Shanghai
```

然后在同一目录下无脑执行：
```bash
docker-compose up -d
```
启动瞬间，图床将在后台以最高效率纯净运转，一切大功告成！

*(PS: 对于仍旧偏爱传统物理文件管理配置的老玩家，系统仍包含“智能回落构建”机制：当检测不到上述环境变量时，只要挂载一个空的 `/app/conf` 数据卷进去，系统便会依然乖乖在里面为你生成默认填报模板。)*

#### 👉 使用原生 Docker 命令行 (纯净一行流)

如果你连文件都不想建，就喜欢那种直接把代码拍在终端上一键起飞的爽感，可以直接复制并在终端运行以下指令：

```bash
docker run -d \
  --name 123pan-imagehost \
  --restart unless-stopped \
  -p 8080:8080 \
  -e PORT=8080 \
  -e CLIENT_ID="【去123pan获取并在引流号内填入】" \
  -e CLIENT_SECRET="【去123pan获取并在引号内填入】" \
  -e PARENT_FILE_ID="" \
  -e API_TOKEN="PRIVATE_123_KEY" \
  -e TZ="Asia/Shanghai" \
  你的用户名/123pan-imagehost:latest
```

---

## 🛠️ 第三方截图工具接入 (ShareX 等)

本图床严格遵守 `RESTful API` 规范，对第三方工具接入极为友好。以下使用 **ShareX** 为例进行配置：

1. 添加 **自定义上传者**；
2. **请求 URL**：`http://修改为你的IP或域名:8080/api/upload` 
3. **HTTP 方法**：`POST`
4. **请求主体配置**：`Multipart/form-data`
5. **文件表单值 / 参数名**：`file`
6. **请求标头 (Headers)**：
   - 增加一行 `Authorization`，值为 `Bearer 这里填你的api_token密码` 
     *(也可以简化为追加到 URL 后面：`?token=你的配置密码`)*
7. **获取 URL 响应路径**：`$json:url$`

配置完毕后，即可享受一键截图瞬移至 123pan 并提取直链的极致体验！
