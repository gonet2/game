FROM golang:latest
MAINTAINER xtaci <daniel820313@gmail.com>
ENV GOBIN /go/bin
COPY src /go/src
WORKDIR /go
RUN go install game
RUN rm -rf pkg src
ENTRYPOINT /go/bin/game
EXPOSE 51000
