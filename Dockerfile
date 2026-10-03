FROM golang:1.25.5-alpine3.23

# community repo is required for yt-dlp & ffmpeg.
#
# yt-dlp alone is not enough for YouTube: it needs the yt-dlp-ejs challenge
# scripts plus a JavaScript runtime, or the n-parameter/signature challenge fails
# and downloads come back as HTTP 403 / "Sign in to confirm you're not a bot".
# Alpine's yt-dlp already depends on yt-dlp-ejs, which pulls in deno (the only
# runtime yt-dlp enables by default), so no explicit JS runtime package is needed
# here — installing nodejs would just sit unused.
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