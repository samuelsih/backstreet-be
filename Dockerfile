FROM golang:1.20-alpine as builder

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /backstreet-api .



FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /backstreet-api /backstreet-api

EXPOSE 8080

ENTRYPOINT [ "/backstreet-api" ]
