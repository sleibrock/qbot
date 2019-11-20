QBot, a bot for queueing
==============================

QBot is designed to queue up users in order and allow the host to pop from the queue to gain one (or more) users for various purposes. It is designed for use with IRC chatrooms or Twitch channels.


In order to use this, you must generate an OAuth code for Twitch by connecting your Twitch account to Twitch Chat OAuth Password Generator. Using that, use your Twitch password and generated OAuth key in a configuration file and run the bot as follows.

```bash
# linux/macos
$ ./qbot your_config.json

# windows
$ ./qbot.exe your_config.json
```
