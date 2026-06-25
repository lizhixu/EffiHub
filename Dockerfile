FROM alpine:latest

WORKDIR /app

# 安装必要工具和非 root 用户
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S appuser && adduser -S appuser -G appuser

# 直接复制 CI 编译好的二进制文件
COPY dist/effihub-linux-amd64 ./effihub
COPY static/ ./static/

RUN chown -R appuser:appuser /app && \
    chmod +x effihub

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./effihub"]
