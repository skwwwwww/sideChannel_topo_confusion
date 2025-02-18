# 第一阶段：构建 Go 程序（保持不变）
FROM golang:1.23.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./ce/main.go

# 第二阶段：使用 Debian 精简镜像
FROM alpine:3.16
RUN apk add --no-cache curl

WORKDIR /app
COPY --from=builder /app/main /app/main
COPY ./config /app/
ENTRYPOINT ["/app/main"]