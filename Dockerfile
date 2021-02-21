FROM golang:1.15.6-alpine3.12 as builder

RUN apk update \
  && apk add --no-cache git curl \
  && go get -u github.com/cosmtrek/air \
  && chmod +x ${GOPATH}/bin/air

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o /main ./cmd

FROM alpine:3.12

COPY --from=builder /main .

ENV PORT=${PORT}
ENTRYPOINT ["/main web"]
