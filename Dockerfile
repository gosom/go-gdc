FROM golang:1.16.5-alpine as builder

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go install ./...

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /usr/bin

COPY --from=builder /go/bin .

ENTRYPOINT ["go-gdc", "api"]
