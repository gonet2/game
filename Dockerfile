FROM golang:1.4
MAINTAINER xtaci <daniel820313@gmail.com>
ENV GOBIN /go/bin
COPY . /go
WORKDIR /go
RUN wget -qO- https://raw.githubusercontent.com/pote/gpm/v1.3.2/bin/gpm | bash
RUN go install game_server
ENTRYPOINT /go/startup.sh
EXPOSE 8800
