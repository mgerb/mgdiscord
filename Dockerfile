FROM golang:1.14.4-alpine3.12

RUN apk add --no-cache git alpine-sdk pkgconfig opus-dev opusfile-dev

WORKDIR /go/src/github.com/mgerb/mgdiscord

ADD . .
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o /build/bot .


FROM jrottenberg/ffmpeg:4.1-alpine

RUN apk update
RUN apk add --no-cache ca-certificates opus-dev opusfile-dev

# install python for youtube-dl
RUN apk add python3
RUN ln -s /usr/bin/python3 /usr/bin/python & \
  ln -s /usr/bin/pip3 /usr/bin/pip

WORKDIR /bot

COPY --from=0 /build /server

ENTRYPOINT ["/server/bot"]
