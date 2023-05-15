FROM golang:1.20.4  AS builder
WORKDIR /go/build
COPY ./cmd/ ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o pod-exporter .

FROM alpine:3.16.0  
WORKDIR /root/
COPY --from=builder /go/build/pod-exporter ./
CMD ["./pod-exporter"]