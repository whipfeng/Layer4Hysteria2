FROM rockylinux:8

RUN dnf install -y \
    gcc \
    glibc-devel \
    wget \
    tar

# 安装 Go 1.23
WORKDIR /usr/local
RUN wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz && \
    tar -xzf go1.23.0.linux-amd64.tar.gz

ENV PATH=/usr/local/go/bin:$PATH
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /build



# 使用官方 Go 镜像，并指定版本为 1.23
#FROM golang:1.23

# 安装交叉编译所需的工具
#RUN apt-get update && apt-get install -y \
#    gcc-x86-64-linux-gnu \
#    g++-x86-64-linux-gnu \
#    libc6-dev-i386

# 设置工作目录
#WORKDIR /go/src

# 设置 Go 环境
#ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc

# 进行编译
#RUN go build -o liblayer4hysteria2.so -buildmode=c-shared .