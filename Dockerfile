FROM golang:1.17-alpine AS builder

WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o finwise .

# 使用alpine作为最终镜像
FROM alpine:latest

WORKDIR /app

# 安装必要的依赖
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制编译好的应用和配置文件
COPY --from=builder /app/finwise /app/
COPY --from=builder /app/conf /app/conf
COPY --from=builder /app/static /app/static
COPY --from=builder /app/views /app/views

# 创建日志目录
RUN mkdir -p /app/logs

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["/app/finwise"] 