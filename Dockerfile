FROM golang

ADD . /server

WORKDIR "/server"

RUN go build server.go

ENTRYPOINT /server/server

EXPOSE 8080
