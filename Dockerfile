# Stage 1: Build
FROM golang:1.25-alpine AS builder
# Instal gcc untuk SQLite
RUN apk add --no-cache build-base 
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o bmi-app .

# Stage 2: Run
FROM alpine:latest
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Ambil binary-nya
COPY --from=builder /app/bmi-app .

# AMBIL JUGA file HTML-nya (Ini yang tadi kurang!)
COPY --from=builder /app/index.html .

EXPOSE 8080
CMD ["./bmi-app"]