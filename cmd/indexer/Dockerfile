FROM golang:1.22 AS builder

# Build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 go build -o indexer ./cmd/indexer

# Copy binary into slim image
FROM alpine
WORKDIR app
COPY --from=builder /src/indexer .
CMD ["/app/indexer"]