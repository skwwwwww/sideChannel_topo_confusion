# 使用 Golang 官方镜像
FROM golang:1.22.5

# 设置工作目录
WORKDIR /app

# 将 go.mod 和 go.sum 复制到容器中
COPY go.mod go.sum ./

# 下载依赖（go mod）
RUN go mod tidy

# 将源代码复制到容器中
COPY . .

# 构建 Go 程序
RUN go build -o myapp ./ce/criticalpath/criticalPath.go

# 设置容器启动时执行的命令
CMD ["./myapp"]
