FROM golang:1.17.0-alpine3.13 as builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --update make

WORKDIR /home

COPY . /home/

RUN ls -l && go env -w GO111MODULE=on  \
    && go env -w GOPROXY="https://goproxy.cn,direct" \
    && make vendor \
    && make linux_build


FROM alpine:3.13.6 as package

MAINTAINER ntptune

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --update tzdata

ENV TZ=Asia/Shanghai

RUN echo "Asia/Shanghai" > /etc/timezone &&\
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ARG TAG_NAME

LABEL maintainer="${TAG_NAME}"

ENV TZ=Asia/Shanghai

WORKDIR /home

ADD config /home/config
COPY --from=0  /home/app  /home/

CMD ["./app"]
