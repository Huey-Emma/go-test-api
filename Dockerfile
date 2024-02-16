FROM golang:1.22-alpine AS builder 

WORKDIR /app

COPY go.mod go.sum .

RUN go get -d -v ./...

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o api .

FROM alpine:edge AS production

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/api api

COPY --from=builder /app/.env .env

EXPOSE 8000

CMD ["./api"]
