FROM docker.io/golang:1.20 as builder

RUN mkdir /app

COPY cmd /app/cmd
COPY sql /app/sql
COPY templates /app/templates
COPY pkg /app/pkg
COPY go.mod go.sum /app/

WORKDIR /app

RUN go mod tidy

RUN CGO_ENABLED=0 go build -o goWebAppTesting ./cmd/web

RUN CGO_ENABLED=0 go build -o goAPITesting ./cmd/api


RUN chmod +x /app/goWebAppTesting

RUN chmod +x /app/goAPITesting

FROM alpine:latest

RUN mkdir /app

WORKDIR /app

RUN mkdir -p /app/static/img

COPY --from=builder /app /app

CMD [ "/app/goWebAppTesting"]

