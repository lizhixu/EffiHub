FROM --platform=linux/amd64 golang:1.21-alpine AS builder

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod ./
RUN go mod download

# 再复制源代码
COPY main.go ./
COPY config/ ./config/
COPY models/ ./models/
COPY handlers/ ./handlers/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o effihub .

FROM --platform=linux/amd64 alpine:3.19

WORKDIR /app

# 安装必要工具和非 root 用户
RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -u 1000 appuser

COPY --from=builder /build/effihub ./
COPY static/ ./static/

RUN chown -R appuser:appuser /app && \
    chmod +x effihub

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./effihub"]
