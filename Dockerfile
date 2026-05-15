FROM golang:1.21 AS builder

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum* ./
RUN go mod download

# 再复制源代码
COPY main.go ./
COPY config/ ./config/
COPY models/ ./models/
COPY handlers/ ./handlers/

# 针对小内存云端编译环境的优化：
# 1. 限制并发编译数为 1 (-p 1)，极大降低编译期的内存占用
# 2. 调低 GOGC=50 更加积极地回收内存
RUN CGO_ENABLED=0 GOOS=linux GOGC=50 go build -p 1 -ldflags="-s -w" -o effihub .

FROM alpine:latest

WORKDIR /app

# 安装必要工具和非 root 用户
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S appuser && adduser -S appuser -G appuser

COPY --from=builder /build/effihub ./
COPY static/ ./static/

RUN chown -R appuser:appuser /app && \
    chmod +x effihub

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./effihub"]