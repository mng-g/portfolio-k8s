# Use the official Go image as a build stage
FROM golang:1.24.1 AS builder

WORKDIR /app
COPY go-frontend-app /app
WORKDIR /app

RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -o go-frontend-app

# Use a minimal base image for the final container
FROM alpine:latest  
RUN apk --no-cache add ca-certificates && apk add --no-cache libc6-compat

WORKDIR /root/
COPY --from=builder /app/go-frontend-app .
COPY --from=builder /app/static /root/static

EXPOSE 9090
CMD ["./go-frontend-app"]