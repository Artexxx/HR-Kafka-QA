FROM golang:1.24-alpine

WORKDIR /app

RUN apk update && apk upgrade && apk add --update alpine-sdk && \
    apk add --no-cache bash git openssh make cmake

EXPOSE 8080
