# syntax=docker/dockerfile:1

# 阶段 1: 编译
FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o effihub .

# 阶段 2: 运行时
FROM scratch

COPY --from=builder /build/effihub /effihub
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY static/ /static/

EXPOSE 8080

ENTRYPOINT ["/effihub"]
