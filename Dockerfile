FROM golang:alpine AS builder
WORKDIR /app
ENV GOPROXY=https://proxy.golang.org CGO_ENABLED=0
COPY . .
RUN go build -ldflags='-s -w' -o main

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main .
CMD ["./main"]