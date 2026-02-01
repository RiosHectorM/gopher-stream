FROM golang:alpine

WORKDIR /app

# Instalamos git por si alguna dependencia lo necesita
RUN apk add --no-cache git

# Copiamos solo el mod primero (el asterisco ayuda si no hay sum)
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Compilamos
RUN go build -o main ./cmd/api/main.go

EXPOSE 8080

CMD ["./main"]