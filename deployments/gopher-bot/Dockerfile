# syntax=docker/dockerfile:1

FROM golang:1.22.7

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.56.2

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY internal internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /gopher-bot cmd/main.go

EXPOSE 8080

CMD [ "/gopher-bot" ]
