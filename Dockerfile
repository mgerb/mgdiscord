FROM golang:1.11.12-alpine3.10

RUN apk add --no-cache git alpine-sdk pkgconfig opus-dev opusfile-dev

WORKDIR /go/src/github.com/mgerb/mgdiscord

ADD . .
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o /build/bot .


FROM wernight/youtube-dl

RUN apk update
RUN apk add --no-cache ca-certificates opus-dev opusfile-dev

WORKDIR /bot

COPY --from=0 /build /server

ENTRYPOINT ["/server/bot"]
