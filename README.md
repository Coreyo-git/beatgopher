# BeatGopher - Discord Music Bot

BeatGopher is a Discord music bot written in Go, designed to play music in your Discord server.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.18 or later)
- A Discord Bot Token. You can create one [here](https://discord.com/developers/applications).

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/your-username/BeatGopher.git
    cd BeatGopher
    ```

2.  **Install dependencies:**
    ```sh
    go mod tidy
    ```

3.  **Configuration:**
    - Rename `src/example.env` to `src/.env`.
    - Add your Discord bot token and other settings to `src/.env`.

## Usage

1.  **Run the bot:**
    ```sh
    go run src/main.go
    ```

2.  **Invite the bot to your server.**

3.  **Use commands in your Discord server:**
    - `!play <song_name_or_url>`: Plays a song.
    - `!skip`: Skips the current song.
    - `!stop`: Stops the music and clears the queue.
    - `!queue`: Shows the current music queue.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
