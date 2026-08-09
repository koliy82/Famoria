FROM golang:1.25.5-alpine3.23

# community repo is required for yt-dlp & ffmpeg.
# nodejs is needed as a JS runtime: YouTube's bot-detection ("Sign in to confirm
# you're not a bot") requires yt-dlp to evaluate the n-challenge JS, which fails
# silently without a JS engine — especially from datacenter IPs.
RUN echo "https://dl-cdn.alpinelinux.org/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && apk upgrade && \
    apk add --no-cache bash git openssh ffmpeg yt-dlp nodejs

LABEL maintainer="Koliy82 <rutopruter@gmail.com>"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bot cmd/main.go

CMD ["./bot"]