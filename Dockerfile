FROM golang:1.23-alpine AS builder

WORKDIR /build
ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -o main main.go

FROM alpine:3.20.3
WORKDIR /build

COPY --from=builder /build/main /build/main

ENV LISTEN="localhost:25565"
ENV TARGET="remotehost:25565"

CMD ["./main"]