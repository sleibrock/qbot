package internal 

import (
	"time"
)


// Information needed to join an IRC server/channel
type Settings struct {
	Name     string `json:"name"`
	Channel  string `json:"channel"`
	Password string `json:"password"`
	Port     string `json:"port"`
	Server   string `json:"server"`
}

func DefaultSettings() Settings {
	return Settings{
		Name: "your_twitch_name",
		Channel: "your_twitch_channel",
		Password: "your_oauth_password_here",
		Port: "6667",
		Server: "irc.chat.twitch.tv",
	}
}



// A message struct containing a Username and a message string (Contents)
type Message struct {
	Name      string
	Contents  string
	TimeRecv  time.Time
}

// Quick wrapper to create a new message with a current time
func NewMsg(name string, msg string) Message {
	return Message{
		Name: name,
		Contents: msg,
		TimeRecv: time.Now(),
	}
}

func (m *Message) ToPlayer() Player {
	return Player{
		Name: m.Name,
		TimeJoined: time.Now(),
	}
}

// A struct representing a Player and the time they joined the queue
type Player struct {
	Name       string
	TimeJoined time.Time 
}



// end structs.go
