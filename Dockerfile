# Dockerfile for maven-proxy

# 构建阶段
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o maven-proxy cmd/maven-proxy/main.go

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates

# 创建应用目录
WORKDIR /root

# 从构建阶段复制二进制文件
COPY --from=builder /app/maven-proxy .

# 复制配置文件
COPY --from=builder /app/config.yaml .

# 创建数据目录
RUN mkdir -p /data/data /data/log

# 暴露端口
EXPOSE 8880

# 启动命令
CMD ["./maven-proxy", "-c", "config.yaml"]
