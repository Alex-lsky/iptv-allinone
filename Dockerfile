# 使用官方 Go 运行时作为父镜像
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 和 sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o main .

# 使用最小的 Alpine 镜像作为最终阶段
FROM alpine:latest

# 安装 ca-certificates 以支持 HTTPS 请求
RUN apk --no-cache add ca-certificates

# 创建非 root 用户
RUN adduser -D -s /bin/sh appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件和配置文件
COPY --from=builder /app/main .
COPY --from=builder /app/config.json .

# 更改文件所有者
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 35455

# 运行应用
CMD ["./main"]