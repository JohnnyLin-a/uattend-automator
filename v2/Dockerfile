FROM golang:16-alpine as builder
# build golang app


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