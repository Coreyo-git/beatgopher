# ---- Builder Stage ----
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache ca-certificates ffmpeg curl python3 git opus-dev gcc build-base

RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Production build
ENV CGO_CFLAGS="-O2"
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /beatgopher ./main.go

# ---- Test Stage ----
FROM builder AS test
ENV CGO_ENABLED=1
CMD ["go", "test", "-v", "./..."]

# ---- Debug Stage ----
FROM builder AS debug
# Install Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Ensure CGO is enabled debug build
ENV CGO_ENABLED=1
EXPOSE 2345

CMD ["dlv", "debug", "--listen=:2345", "--headless=true", "--api-version=2", "main.go"]

# ---- Final Stage ----
FROM alpine:latest AS release
RUN apk add --no-cache ca-certificates ffmpeg curl python3 opus
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp
COPY --from=builder /beatgopher /beatgopher
CMD ["/beatgopher"]