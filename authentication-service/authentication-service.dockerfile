#builder image
FROM golang:1.18-alpine AS BUILDER
WORKDIR /app

COPY . .

RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

RUN CGO_ENABLED=0 go build -o auth_app ./cmd/api

RUN chmod +x ./auth_app

#tiny image
FROM alpine:latest
WORKDIR /app

COPY --from=BUILDER /app/auth_app .
COPY --from=builder /app/migrate ./migrate
COPY ./db/migrations ./migrations
COPY start.sh .
COPY wait-for.sh .