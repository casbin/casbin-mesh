FROM golang:1.17.1-alpine as builder
WORKDIR /root
COPY ./go.mod ./
COPY ./go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download
COPY . .
RUN export GO111MODULE=on && CGO_ENABLED=0 GOOS=linux go build  -ldflags "-s -w" -o build/casmesh cmd/app/*.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/' /etc/apk/repositories \
    && apk update -v \
    && apk add bash \
    && apk upgrade -v --no-cache

WORKDIR /root
COPY --from=builder /root/build/casmesh ./

RUN mkdir -p /casmesh/file

VOLUME /casmesh/data

COPY ./docker-entrypoint.sh ./
RUN chmod +x /root/docker-entrypoint.sh
ENTRYPOINT ["/root/docker-entrypoint.sh"]

EXPOSE 4002

CMD ["/root/casmesh", "-raft-address", "0.0.0.0:4002", "/casmesh/data/data"]
