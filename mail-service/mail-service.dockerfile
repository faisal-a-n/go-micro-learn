#builder image
FROM golang:1.18-alpine AS BUILDER

RUN mkdir /app
COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 go build -o mailer_app ./cmd/api
RUN chmod +x /app/mailer_app

#tiny image
FROM alpine:latest
RUN mkdir /app
WORKDIR /app

COPY --from=BUILDER /app/templates ./templates
COPY --from=BUILDER /app/mailer_app .

CMD [ "./mailer_app" ]