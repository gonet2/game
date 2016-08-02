FROM golang:latest
MAINTAINER xtaci <daniel820313@gmail.com>
COPY . /go/src/game
RUN go install game
ENTRYPOINT /go/bin/game
EXPOSE 51000
