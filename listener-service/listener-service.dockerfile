#builder image
FROM golang:1.18-alpine AS BUILDER

RUN mkdir /app
COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 go build -o listener_app .

RUN chmod +x /app/listener_app

#tiny image
FROM alpine:latest
RUN mkdir /app
COPY --from=BUILDER /app/listener_app /app

CMD [ "/app/listener_app" ]