# JacuzziBot
A fun Discord bot with the features described in [`utils\help.go`](https://github.com/janimationd/JacuzziBot/blob/main/utils/help.go).

The code is organized as follows:
- `constants/`: constant values
- `db/`: database-related functionality. Stores persistent data on disk using [the bbolt library](https://github.com/etcd-io/bbolt) for later fetching.
- `discord/`: logic specific to Discord. This bot might eventually evolve to interact with multiple services/APIs so this is where anything Discord-specific lives. The bot listens and responds to events from the Discord API by utilizing [the discordgo library](https://github.com/bwmarrin/discordgo).
- `models/`: model object types
- `workflows/`: these are the "features" of the bot. After triggers have arrived we execute workflows which contain the "since this happened, do this" logic. Ideally these workflows are agnostic to the models of any specific service or API.

## Setup
1. To install the bot in a new server, visit the Install Link provided in the Discord developer portal. (Johnathan has this)
1. Install Go >= 1.25.5: https://go.dev/doc/install.
1. Set the `DISCORD_TOKEN` environment variable to the app's Discord auth token. (Johnathan has this)
1. `go run .` to both compile and run the app executable locally.
1. `Ctrl+C` to stop.

Example local output:
```
...\JacuzziBot> go run .
2025/12/29 21:19:39 JacuzziBot is now running. Press CTRL-C to exit.
2025/12/29 21:19:39 Discord session opened and waiting for events.
2025/12/29 21:19:51 Incoming message by 123456789012345678 (Sally): Hi everyone!
2025/12/29 21:19:51 User 123456789012345678 (Sally) gained 10.00 points, for a total of 20.00 points.
2025/12/29 21:19:58 JacuzziBot shutting down.
2025/12/29 21:19:58 Discord session closed.
```

## License & Project Philosophy

This project is source-available, not open source.

You are welcome to:
- Read and learn from the code
- Contribute improvements
- Run the bot privately for testing or development

You may not:
- Use this code in a commercial product
- Clone or lift features for use in a different Discord bot

The goal is to keep this a fun, collaborative project among friends while retaining primary control over the bot itself and its future direction.

If you have an idea that might fall into a gray area, just ask — permission is often easier than rewriting code 🙂
