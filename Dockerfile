FROM golang:alpine AS builder
WORKDIR /build
ADD go.mod .
COPY . .
EXPOSE 8080
RUN go build -o server .
FROM alpine
WORKDIR /build
COPY --from=builder /build/server /build/server
CMD ["./server"]