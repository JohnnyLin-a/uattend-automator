FROM golang:1.17-buster AS builder
# build golang app
COPY ./v2 /root/src/v2/
WORKDIR /root/src/v2/
RUN go build ./cmd/main/main.go

FROM ubuntu:20.04
ENV TZ="America/New_York" DEBIAN_FRONTEND="noninteractive"
WORKDIR /root

# Set TZ, install firefox & driver
RUN cd /root && echo $TZ > /etc/timezone && \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y tzdata firefox firefox-geckodriver && \
    rm /etc/localtime && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    dpkg-reconfigure -f noninteractive tzdata && \
    rm -rf /var/lib/apt/lists/*

ADD https://github.com/SeleniumHQ/selenium/releases/download/selenium-3.141.59/selenium-server-standalone-3.141.59.jar /root/vendor/

COPY --from=builder /root/src/v2/main /root/