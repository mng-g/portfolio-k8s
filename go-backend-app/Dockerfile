# Build stage: compile the Go backend app
FROM golang:1.24.1 AS builder
WORKDIR /app
# Copy the backend source code
COPY . .
RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -o go-backend-app

# Final stage: minimal runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates libc6-compat
WORKDIR /root/
COPY --from=builder /app/go-backend-app .
EXPOSE 9191
CMD ["./go-backend-app"]
