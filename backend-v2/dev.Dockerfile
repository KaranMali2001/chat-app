FROM golang:1.24-alpine

#

# Install air
RUN go install github.com/air-verse/air@latest

# Add Go bin to path (where air gets installed)
ENV PATH="/go/bin:$PATH"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air"]
