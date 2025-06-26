FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./


RUN go mod download

COPY . /app/

RUN go build -o main ./main.go
##Stage 2
FROM alpine

WORKDIR /app

COPY --from=build /app/main .
COPY .env . 
EXPOSE 8080

CMD [ "./main"]