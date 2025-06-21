FROM golang:latest AS builder

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code and build the application
COPY . ./
# Build a statically linked binary with CGO disabled
RUN CGO_ENABLED=0 go build -o /kubernetes-mcp-server cmd/kubernetes-mcp-server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /kubernetes-mcp-server /kubernetes-mcp-server
EXPOSE 8081

ENTRYPOINT ["/kubernetes-mcp-server", "--sse-port", "8081", "--log-level", "2"]



