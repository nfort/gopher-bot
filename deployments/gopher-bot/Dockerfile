# syntax=docker/dockerfile:1

FROM golang:1.22.7

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY internal internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /gopher-bot cmd/main.go

EXPOSE 8080

CMD [ "/gopher-bot" ]
