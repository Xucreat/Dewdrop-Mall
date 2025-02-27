FROM golang:1.23.0 as builder

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY . .

# 调试信息：查看 /app 目录的内容
RUN ls -la /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -ldflags="-w -s" -o main ./cmd
RUN mkdir publish  \
    && cp main publish  \
    && cp -r conf publish

FROM busybox:1.28.4

WORKDIR /app

COPY --from=builder /app/publish .

# 指定运行时环境变量
ENV GIN_MODE=release
EXPOSE 3000

ENTRYPOINT ["./main"]