# 使用多阶段构建减小镜像体积
# 第一阶段：构建 Go 程序
FROM golang:1.23.0 AS builder

# 设置工作目录
WORKDIR /app

# 先拷贝依赖定义文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 拷贝整个项目
COPY . .

# 构建可执行文件（注意 main.go 在 ce 目录下）
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./ce/main.go

# 第二阶段：构建最小运行镜像
FROM alpine:3.18

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/main /app/main

COPY . .

# 设置容器启动命令
ENTRYPOINT ["/app/main"]