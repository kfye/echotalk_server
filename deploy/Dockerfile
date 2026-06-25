# 多阶段构建：编译期用完整工具链，运行期用精简镜像。
FROM golang:1.26-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/echotalk ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/echotalk /app/echotalk
COPY configs/config.example.yaml /app/configs/config.yaml
EXPOSE 8080
ENTRYPOINT ["/app/echotalk", "-config", "/app/configs/config.yaml"]
