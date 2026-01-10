# BeatGopher - Discord Music Bot

BeatGopher is a Discord music bot written in Go, designed to play music from YouTube in your Discord server with queue management and playlist support.

## Features

- YouTube Integration: Play songs directly from YouTube URLs or search by name
- Queue Management: Add, view, remove, and manage your music queue with pagination
- Playlist Support: Load entire YouTube playlists with customizable options
- Randomization: Shuffle playlist songs for variety
- Slash Commands: Modern Discord slash command interface
- Docker Support: Easy deployment with Docker containers

## Getting Started

### Prerequisites

**For Docker setup (recommended):**
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- A Discord Bot Token ([create one here](https://discord.com/developers/applications))

**For local development:**
- [Go](https://golang.org/doc/install) (version 1.25 or later)
- [FFmpeg](https://ffmpeg.org/download.html) - Required for audio processing
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - YouTube downloader
- Opus development libraries (for audio encoding)
- GCC/build tools (for CGO)
- A Discord Bot Token

#### Installing System Dependencies (Local Development)

> Skip this section if you're using Docker - all dependencies are handled automatically in the container.

**Ubuntu/Debian:**
```sh
sudo apt update
sudo apt install ffmpeg libopus-dev build-essential
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
chmod a+rx /usr/local/bin/yt-dlp
```

**macOS:**
```sh
brew install ffmpeg opus yt-dlp
```

**Alpine Linux:**
```sh
apk add ffmpeg opus-dev gcc build-base python3
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
chmod a+rx /usr/local/bin/yt-dlp
```

### Installation

#### Docker (Recommended)

1. **Clone the repository:**
   ```sh
   git clone https://github.com/coreyo-git/beatgopher.git
   cd beatgopher
   ```

2. **Configuration:**
   ```sh
   cp example.env .env
   # Edit .env and add your Discord bot token
   ```

3. **Run (Production):**
   ```sh
   docker build --target release -t beatgopher .
   docker run -d --name beatgopher --env-file .env beatgopher
   ```

#### Local Development

1. **Clone and install dependencies:**
   ```sh
   git clone https://github.com/coreyo-git/beatgopher.git
   cd beatgopher
   go mod tidy
   ```

2. **Configuration:**
   ```sh
   cp example.env .env
   # Edit .env and add your Discord bot token
   ```

3. **Run:**
   ```sh
   CGO_ENABLED=1 go run main.go
   ```

## Discord Commands

| Command | Description |
|---------|-------------|
| `/play <query>` | Play a song from YouTube URL or search term |
| `/playlist <url> [total] [random]` | Add songs from a YouTube playlist |
| `/skip` | Skip the current song |
| `/stop` | Stop playback and clear the queue |
| `/showqueue [page]` | Display the current music queue (10 songs per page) |
| `/remove [position] [query]` | Remove a song by position number or title search |

### Examples

```
/play Never Gonna Give You Up
/play https://www.youtube.com/watch?v=dQw4w9WgXcQ
/playlist https://www.youtube.com/playlist?list=PLExample total:50 random:true
/showqueue page:2
/remove position:3
/remove query:rickroll
```

## Development

### Project Structure

```
beatgopher/
├── commands/           # Discord slash command handlers
├── config/             # Configuration management
├── discord/            # Discord session and voice handling
├── player/             # Music player and audio streaming
├── queue/              # Queue management
├── services/           # External services (YouTube, FFmpeg)
├── mocks/              # Test mocks
├── main.go             # Entry point
├── Dockerfile          # Multi-stage Docker build
└── docker-compose.yaml # Development environment
```

### Docker Build Stages

The Dockerfile provides multiple build targets:

| Target | Purpose |
|--------|---------|
| `builder` | Compiles the application with CGO |
| `test` | Runs the test suite |
| `debug` | Development with Delve debugger |
| `release` | Minimal production image |

**Run tests:**
```sh
docker build --target test -t beatgopher-test . && docker run --rm beatgopher-test
```

**Debug with Delve:**
```sh
docker-compose up dev
# Delve listens on port 2345
```

### Debugging with VS Code

The debug stage runs Delve on port 2345. Add this to your `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Attach to Docker",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "port": 2345,
      "host": "127.0.0.1"
    }
  ]
}
```

### Running Tests Locally

Tests require CGO and Opus libraries. If you have the dependencies installed:

```sh
CGO_ENABLED=1 go test -v ./...
```

Otherwise, use the Docker test stage.

### Adding New Commands

1. Create a new file in `commands/` directory
2. Implement the command handler function
3. Register the command in `init()` using the `Commands` map
4. The bot automatically registers commands on startup

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see [LICENSE.md](LICENSE.md) for details.

## Acknowledgments

- [DiscordGo](https://github.com/bwmarrin/discordgo) - Discord API wrapper
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - YouTube media extraction
- [FFmpeg](https://ffmpeg.org/) - Audio processing
- [gopus](https://github.com/layeh/gopus) - Opus audio encoding
