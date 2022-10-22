#builder image
FROM golang:1.18-alpine AS BUILDER

RUN mkdir /app
COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 go build -o logger_app ./cmd/api

RUN chmod +x /app/logger_app

#tiny image
FROM alpine:latest
RUN mkdir /app
COPY --from=BUILDER /app/logger_app /app

CMD [ "/app/logger_app" ]