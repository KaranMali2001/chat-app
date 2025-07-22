# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Stage 2: Run
FROM alpine

WORKDIR /app

COPY --from=builder /app/main .
COPY .env.prod .
EXPOSE 8080

CMD ["./main"]
