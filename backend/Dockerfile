FROM golang:1.24 as builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
# We use CGO_ENABLED=0 to ensure a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -v -o server ./cmd/api/main.go

# Final lightweight stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

# Expose the dynamically managed port
EXPOSE 8080

# Run the server
CMD ["./server"]
