FROM golang:1.25.5-alpine3.23

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

LABEL maintainer="Koliy82 <rutopruter@gmail.com>"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bot cmd/main.go

CMD ["./bot"]