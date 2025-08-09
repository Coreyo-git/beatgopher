# BeatGopher - Discord Music Bot

BeatGopher is a Discord music bot written in Go, designed to play music from YouTube in your Discord server with queue management and playlist support.

## ✨ Features

- 🎵 **YouTube Integration**: Play songs directly from YouTube URLs or search by name
- 📋 **Queue Management**: Add, view, and manage your music queue with pagination
- 🎶 **Playlist Support**: Load entire YouTube playlists with customizable options
- 🔀 **Randomization**: Shuffle playlist songs for variety
- 🎛️ **Slash Commands**: Modern Discord slash command interface
- 🐳 **Docker Support**: Easy deployment with Docker containers

## 🚀 Getting Started

### Prerequisites

**For Docker setup:**
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- A Discord Bot Token. You can create one [here](https://discord.com/developers/applications)

**For local development only:**
- [Go](https://golang.org/doc/install) (version 1.24.4 or later)
- [FFmpeg](https://ffmpeg.org/download.html) - Required for audio processing
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - YouTube downloader
- Opus development libraries (for audio encoding)
- A Discord Bot Token. You can create one [here](https://discord.com/developers/applications)

#### Installing System Dependencies (Local Development Only)

> **Note:** Skip this section if you're using Docker - all dependencies are handled automatically in the container.

**Ubuntu/Debian:**
```sh
sudo apt update
sudo apt install ffmpeg libopus-dev
# Install yt-dlp
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
chmod a+rx /usr/local/bin/yt-dlp
```

**macOS:**
```sh
brew install ffmpeg opus yt-dlp
```

**Alpine Linux:**
```sh
apk add ffmpeg opus-dev python3
# Install yt-dlp
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
chmod a+rx /usr/local/bin/yt-dlp
```

### Installation

#### Method 1: Local Development

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/coreyo-git/beatgopher.git
    cd beatgopher
    ```

2.  **Install dependencies:**
    ```sh
    go mod tidy
    ```

3.  **Configuration:**
    - Copy `example.env` to `.env`:
      ```sh
      cp example.env .env
      ```
    - Edit `.env` and add your Discord bot token:
      ```
      TOKEN=YOUR_DISCORD_BOT_TOKEN
      ```



#### Method 2: Docker (Recommended)

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/coreyo-git/beatgopher.git
    cd beatgopher
    ```

2.  **Configuration:**
    - Copy `example.env` to `.env` and add your bot token

3. **Run with Docker (Production): **
    ```sh
    docker build -t beatgopher .
    
    # run with env var
    docker run -d --name beatgopher -e TOKEN="YOUR_DISCORD_BOT_TOKEN" beatgopher
    
    # env file
    docker run -d --name beatgopher --env-file .env beatgopher
    ```

5.  **Run with Docker Compose (Development): ** 
    ```sh
    # For development
    docker-compose up
    
    # For production
    docker build -t beatgopher .
    docker run --env-file .env beatgopher
    ```

## 🎵 Usage

### Running the Bot

**Local:**
```sh
go run main.go
```

**Docker:**
```sh
docker-compose up dev
```

### Discord Commands

BeatGopher uses modern Discord slash commands:

- `/play <query>` - Play a song from YouTube URL or search term
- `/playlist <url> [total] [random]` - Add songs from a YouTube playlist
  - `total`: Number of songs to add (default: 25)
  - `random`: Randomize playlist order (default: false)
- `/skip` - Skip the current song
- `/stop` - Stop playback and clear the queue
- `/showqueue [page]` - Display the current music queue
  - `page`: View specific page of queue (10 songs per page)

### Examples

```
/play Never Gonna Give You Up
/play https://www.youtube.com/watch?v=dQw4w9WgXcQ
/playlist https://www.youtube.com/playlist?list=PLExample total:50 random:true
/showqueue page:2
```

## 🛠️ Development

### Project Structure

```
beatgopher/
├── commands/          # Discord slash commands
├── config/           # Configuration management
├── discord/          # Discord session wrapper
├── player/           # Music player logic
├── queue/            # Queue management
├── services/         # External services (YouTube, Audio)
├── mocks/            # Test mocks
├── main.go           # Entry point
└── docker-compose.yaml
```

### Development Setup

1. **Install Air for hot reloading:**
   ```sh
   go install github.com/air-verse/air@latest
   ```

2. **Run with hot reload:**
   ```sh
   air
   ```

3. **Run tests:**
   ```sh
   go test ./...
   ```

### Adding New Commands

1. Create a new file in `commands/` directory
2. Implement the command structure with `init()` function
3. Register the command in the `Commands` map
4. The bot will automatically register it on startup

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

For major changes, please open an issue first to discuss what you would like to change.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## 🙏 Acknowledgments

- [DiscordGo](https://github.com/bwmarrin/discordgo) - Discord API wrapper
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - YouTube media extraction
- [FFmpeg](https://ffmpeg.org/) - Audio processing
