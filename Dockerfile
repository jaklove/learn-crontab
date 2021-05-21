FROM golang AS builder
MAINTAINER zhourenjie
WORKDIR /go/src/app
COPY . .
ENV GOPROXY https://goproxy.io
RUN cd /go/src/app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app master/main/main.go
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/app .
EXPOSE 8070
ENTRYPOINT ["./app"]