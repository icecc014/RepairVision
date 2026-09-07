ARG SERVICE
FROM golang:1.22.5-alpine AS builder
ARG SERVICE
WORKDIR /build
COPY services ./services
WORKDIR /build/services/${SERVICE}
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server .

FROM alpine:3.20
ARG SERVICE
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /out/server ./server
COPY --from=builder /build/services/${SERVICE}/etc ./etc
USER app
ENTRYPOINT ["./server"]