# FROM golang:1.17-alpine as dev-env

# WORKDIR /app

# FROM dev-env as build-env
# COPY go.mod /go.sum /app/
# RUN go mod download

# COPY ./*.go /app/

# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /webhook

FROM alpine:3.10 as runtime

COPY ./compile /usr/local/bin/webhook
RUN chmod +x /usr/local/bin/webhook

ENTRYPOINT ["webhook"]