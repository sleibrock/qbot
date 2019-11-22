QBot, a bot for queueing
==============================

QBot is designed to queue up users in order and allow the host to pop from the queue to gain one (or more) users for various purposes. It is designed for use with IRC chatrooms or Twitch channels.


In order to use this, you must generate an OAuth code for Twitch by connecting your Twitch account to Twitch Chat OAuth Password Generator. Using that, use your Twitch password and generated OAuth key in a configuration file and run the bot as follows.

```bash
# linux/macos
$ ./qbot

# windows
$ ./qbot.exe 
```

Upon first execution, you will be given a file `botdata.json` to fill out where you input your Twitch details (username, channel), and must put in an OAuth password generated from [Twitch OAuth Chat Integration](https://twitchapps.com/tmi).

## Commands

```bash
!queue  - show the queue
!join   - join the queue (multiple joins are not allowed)
!leave  - leave the queue

# owner only commands
!pop    - pop a number of players from the queue
```
