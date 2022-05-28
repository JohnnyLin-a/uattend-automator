FROM golang:1.18-alpine AS builder
# build golang app
COPY ./ /root/src/
WORKDIR /root/src/
RUN CGO_ENABLED=0 go build ./cmd/main/main.go

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


COPY --from=builder /root/src/main /root/