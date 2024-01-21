FROM golang:alpine

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

RUN apk add --no-cache git

WORKDIR /opt/project

COPY go.mod go.sum ./

RUN go mod tidy
RUN go build

CMD ["./go-dag"]