FROM golang:1.21

WORKDIR /src

RUN go install github.com/mitranim/gow@latest

CMD ["/go/bin/gow", "run", "./cmd/app"]