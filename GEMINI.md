This is a learning project to create a Discord music bot using Golang. 
The bot will be named BeatGopher and will be able to play music in a Discord server.

Key features to implement:

- Play music from YouTube and other sources.
- Music queueing system.
- Basic playback controls (play, skip, stop).
- Show the current queue.

Tech stack:

- Golang
- discordgo library for Discord API interaction.
- youtube-dl or similar for audio extraction.
- still to be decided

Development rules:
- I will only provide advice and code comments. I will not write or modify code directly.

Command Architecture:
- Slash commands are used (e.g., /play).
- Commands are implemented using a "Command-Centric" approach for modularity.
- Each command (e.g., /play) has its own Go file in `src/commands/`.
- Commands define their `discordgo.ApplicationCommand` (name, description, options) and their `Handler` function.
- An `init()` function within each command file registers the command with a central `commands.Commands` map.
- `main.go` contains an `onReady` handler that iterates through `commands.Commands` to register all defined slash commands with Discord.
- `main.go` also contains an `interactionCreate` handler that dispatches incoming slash command interactions to the appropriate `Handler` function from the `commands.Commands` map.

Folder Structure:
- `src/`: The main application source code directory.
- `src/main.go`: The primary entry point of the application, responsible for Discord session setup, event handlers, and command registration.
- `src/config/`: Contains configuration-related files, primarily `config.go` for loading environment variables and application settings (e.g., bot token).
- `src/commands/`: Contains individual Go files for each slash command (e.g., `play.go`, `skip.go`).
- `src/commands/command.go`: Defines the `Command` struct and the `Commands` map, which acts as the central registry for all commands.