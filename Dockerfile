# ==== 第一阶段: 编译打包层 ====
FROM golang:1.21-alpine AS builder

# 配置国内高速 Go 代理环境，开启跨平台交叉编译所需变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 优先缓存和下载项目依赖
COPY go.mod go.sum ./
RUN go mod download

# 载入完整代码并执行剥离调试信息的极限压缩静态编译 (-ldflags="-w -s")
COPY . .
RUN go build -ldflags="-w -s" -o imagehost ./cmd/main.go

# ==== 第二阶段: 纯净运行环境层 ====
FROM alpine:latest

# 123云盘采用硬 TLS 协议，必须拉取 ca-certificates，否则网络连接拦截报错
# TZdata 用于纠正容器运行内的时间对齐，保障授权握手不过期
RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Shanghai

WORKDIR /app

# 将上一阶段萃取好的精华可执行程序抓取过来
COPY --from=builder /app/imagehost .

# 暴露对外的 HTTP 接口
EXPOSE 8080

# 触发点火
CMD ["./imagehost"]
