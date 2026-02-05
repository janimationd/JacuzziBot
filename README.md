# JacuzziBot
A fun Discord bot with the features described in [`utils\help.go`](https://github.com/janimationd/JacuzziBot/blob/main/utils/help.go).

The code is organized as follows:
- `constants/`: constant values
- `db/`: database-related functionality. Stores persistent data on disk using [the bbolt library](https://github.com/etcd-io/bbolt) for later fetching.
- `discord/`: logic specific to Discord. This bot might eventually evolve to interact with multiple services/APIs so this is where anything Discord-specific lives. The bot listens and responds to events from the Discord API by utilizing [the discordgo library](https://github.com/bwmarrin/discordgo).
- `errors/`: custom error types
- `models/`: model object types
- `scheduler/`: code related to scheduled events
- `utils/`: utility functions
- `workflows/`: these are the "features" of the bot. After triggers have arrived we execute workflows which contain the "since this happened, do this" logic. Ideally these workflows are agnostic to the models of any specific service or API.

## Setup
1. To install the bot for testing in a new server, you'll need to create a new application in [the Discord developer portal](https://discord.com/developers/applications):
    1. Name it whatever you want (e.g. `JacuzziBot Dev<YourName>`).
    1. Under `Installation` tab:
        1. Uncheck `User Install`
        1. Under `Default Install Settings > Guild install`:
            1. Under `Scopes`, add `bot`. `applications.commands` should already be there.
            1. Under `Permissions`, add:
                1. `Add Reactions`
                1. `Attach Files`
                1. `Create Polls`
                1. `Embed Links`
                1. `Mention Everyone`
                1. `Read Message History`
                1. `Send Messages`
                1. `Send Messages in Threads`
                1. `View Channels`
    1. Under `Bot` tab:
        1. Press `Reset Token`, and then copy the new token. Save it somewhere safe (e.g. a password manager).
        1. Under `Authorization Workflow`:
            1. Enable `Server Members Intent`
            1. Enable `Message Content Intent`
            1. Note: it is fine to leave `Public Bot` enabled as long as you don't share your installation link with anyone. Making a bot publicly discoverable is a different process.
1. After you've created and configured your test app, create a new Discord server for testing the bot in.
1. To install the test app in your new server, go to the `Installation` tab in the developer portal for your app.
    1. Open the `Install Link` in your browser. This should redirect you to Discord to install your app in a server of your choice.
1. Install Go >= 1.25.5: https://go.dev/doc/install.
1. Set your `DISCORD_TOKEN` environment variable to your app's Discord auth token that you saved earlier.
    1. If you're developing in VS Code you can permanently store your token in the PowerShell terminal like this: `setx DISCORD_TOKEN "..."`. You'll need to restart VS Code before this change takes effect. Verify with `echo $env:DISCORD_TOKEN`.
1. `go run .` to both compile and run the app executable locally.
1. Once it's running, your client should pick up the commands so you can start using them within about 5 seconds.
1. `Ctrl+C` in your terminal to stop the bot program and exit cleanly.

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

## License

This project is licensed open-source under the MIT License.
