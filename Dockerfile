# Build stage
FROM golang:1.25.1-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main "./cmd"

# Final
FROM alpine:latest

WORKDIR /app

COPY --from=build /app/main .

CMD ["./main"]
