FROM golang:1.25.5-alpine3.23

# community repo is required for yt-dlp & ffmpeg.
RUN echo "https://dl-cdn.alpinelinux.org/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && apk upgrade && \
    apk add --no-cache bash git openssh ffmpeg yt-dlp

LABEL maintainer="Koliy82 <rutopruter@gmail.com>"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bot cmd/main.go

CMD ["./bot"]