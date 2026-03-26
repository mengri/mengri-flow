# ===== 前端构建阶段 =====
FROM node:22-alpine AS frontend

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ .
RUN npm run build

# ===== 后端构建阶段 =====
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制全部源码
COPY . .

# 将前端构建产物覆盖到 web/dist（用于 go:embed）
COPY --from=frontend /app/web/dist ./web/dist

# 编译单一二进制（前端已内嵌）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/mengri-flow ./cmd/server

# ===== 运行阶段（极简镜像） =====
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 只需要一个二进制 + 配置文件，前端已内嵌
COPY --from=builder /app/bin/mengri-flow .
COPY --from=builder /app/config.yaml .

EXPOSE 8080

ENTRYPOINT ["./mengri-flow"]
