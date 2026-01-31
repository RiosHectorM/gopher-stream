FROM golang:1.21-alpine

WORKDIR /app

# Copiamos archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del código
COPY . .

# Compilamos la app
RUN go build -o main ./cmd/api/main.go

EXPOSE 8080

CMD ["./main"]