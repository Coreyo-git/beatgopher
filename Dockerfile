# ---- Builder ----
FROM golang:1.24-alpine AS builder

# Install runtime deps.
# ca-certificates is needed for making HTTPS reqs.
# python3 is needed for yt-dlp.
# git is needed for golang modules
RUN apk add --no-cache ca-certificates ffmpeg curl python3 git opus-dev gcc build-base

# yt-dlp download.
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

# Install air
RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the rest.
COPY . .

# Build
# CGO self container binary
# ldflags makes the binary file smaller
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /beatgopher ./main.go

# ---- Final ----
FROM alpine:latest

# Install runtime deps.
# ca-certificates is needed for making HTTPS reqs.
RUN apk add --no-cache ca-certificates ffmpeg curl python3

# yt-dlp download.
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

# Copy binary from build.
COPY --from=builder /beatgopher /beatgopher

# Copy the .env file.
# Check out docker secrets?
COPY ./.env ./.env

CMD ["/beatgopher"]