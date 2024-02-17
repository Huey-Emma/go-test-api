FROM golang:1.22-alpine AS builder 

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o api .

FROM alpine:edge AS production

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/api api

EXPOSE 8000

CMD ["./api"]
