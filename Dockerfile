FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o messenger .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/messenger .

VOLUME /app/data
ENV MESSENGER_DB=/app/data/messenger.db

EXPOSE 8080
CMD ["./messenger"]
