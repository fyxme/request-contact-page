FROM golang

# Copy the local package files to the container's workspace.
ADD . /server

WORKDIR "/server"

RUN go build server.go

# Run the outyet command by default when the container starts.
ENTRYPOINT /server/server

EXPOSE 8080