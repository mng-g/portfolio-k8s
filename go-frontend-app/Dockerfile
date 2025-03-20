# Build stage: compile the Go frontend app
FROM golang:1.24.1 AS builder
WORKDIR /app
# Copy everything in the go-frontend-app folder into the image
COPY . .
RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -o go-frontend-app

# Final stage: use a minimal image
FROM alpine:latest
RUN apk --no-cache add ca-certificates libc6-compat
WORKDIR /root/
# Copy the compiled binary and the static folder
COPY --from=builder /app/go-frontend-app .
COPY --from=builder /app/static /root/static
EXPOSE 9090
CMD ["./go-frontend-app"]
