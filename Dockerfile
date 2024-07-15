FROM golang:1.21

WORKDIR /go/project/chartService

ADD go.mod go.sum main.go ./
ADD internal ./internal
ADD migrations ./migrations

CMD ["go", "run", "main.go"]