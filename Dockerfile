# ---- Builder ----
FROM golang:1.24-alpine AS builder

# needed to fetch Go modules.
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the rest.
COPY . .

# Build
# CGO self container binary
# ldflags makes the binary file smaller
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /beatgopher ./src/main.go

# ---- Final ----
FROM alpine:latest

# Install runtime deps.
# ca-certificates is needed for making HTTPS reqs.
RUN apk add --no-cache ca-certificates ffmpeg curl

# yt-dlp download.
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

# Copy binary from build.
COPY --from=builder /beatgopher /beatgopher

# Copy the .env file.
# Check out docker secrets?
COPY ./.env ./.env

CMD ["/beatgopher"]