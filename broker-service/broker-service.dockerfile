#builder image
FROM golang:1.18-alpine AS BUILDER

RUN mkdir /app
COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 go build -o broker_app ./cmd/api

RUN chmod +x /app/broker_app

#tiny image
FROM alpine:latest
RUN mkdir /app
COPY --from=BUILDER /app/broker_app /app

CMD [ "/app/broker_app" ]