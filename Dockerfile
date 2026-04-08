# ==== 第一阶段: 编译打包层 ====
FROM golang:alpine AS builder

# 第一时间安装最详尽的根证书与时域数据源，这是最后提取所必须的物质
RUN apk add --no-cache ca-certificates tzdata

# 开启 CGO_ENABLED=0，这是逃逸出系统依赖，实现真正的静态单文件极限打包的关键
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-w -s" -o imagehost ./cmd/main.go

# ==== 第二阶段: 绝对物理真空层 (Scratch) ====
# 既然 Alpine 的组件总能被扫出各种各样的零日陈旧漏洞，那么最好的防御就是不要任何组件系统！
# Scratch 中连 Shell / Bash / busybox 都没有，真正从物理维度杜绝了各类越权注入与 CVE 扫描报警。
FROM scratch

# 从刚才的工厂层提炼出我们的核心支撑数据（时间树与 HTTPS 证书）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/imagehost .

EXPOSE 8080

CMD ["./imagehost"]
