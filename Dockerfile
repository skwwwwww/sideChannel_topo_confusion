# 第一阶段：构建阶段
FROM golang:1.23.0 AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod tidy

# 复制源代码
COPY . .

# 构建应用程序
RUN go build -o myapp ./ce/main.go && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /root/.cache/go-build

# 第二阶段：运行时阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/myapp .

# 设置容器启动时执行的命令
CMD ["./myapp"]