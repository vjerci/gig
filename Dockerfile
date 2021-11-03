FROM golang:1.17.2-alpine3.14

RUN apk add g++ make curl && \
    mkdir -p /go/src/app

WORKDIR /go/src/app

COPY . .

RUN chmod u+x wait-for-it.sh
