package main


import (
	"fmt"
	"net"
	"net/textproto"
	"encoding/json"
	"strings"
	"regexp"
	"errors"
	"log"
	"os"
	//"io"
	"io/ioutil"
	"bufio"
	"time"
	"container/list"
)

// Message parsing and date stuff
const ESTFormat = "Jan 2 15:04:05 EST"
var MsgRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)

var CmdRegex *regexp.Regexp = regexp.MustCompile(`^!(\w+)\s?(\w+)?`)

/*
//deprecated, removing later
type OAuthCred struct {
	Password string `json:"password,omitempty"`
	ClientID string `json:"client_id,omitempty"`
}
*/


// Information needed to join an IRC server/channel
type Settings struct {
	Name     string `json:"name"`
	Channel  string `json:"channel"`
	Password string `json:"password"`
	Port     string `json:"port"`
	Server   string `json:"server"`
}

// Our QBot, who stores a queue of players
type QBot struct {
	// actual tcp connection object 
	conn           net.Conn

	// bot settings/data read from external file
	Config         *Settings

	// Bot timers
	MsgRate        time.Duration
	StartTime      time.Time

	// fields related to the queue itself
	MaxSize        int 
	queue          *list.List
}

// A message struct containing a Username and a message string (Contents)
type Message struct {
	Name      string
	Contents  string
	TimeRecv  time.Time
}

// A struct representing a Player and the time they joined the queue
type Player struct {
	Name string
	TimeJoined string
}


// Quick wrapper to create a new message with a current time
func NewMsg(name string, msg string) Message {
	return Message{
		Name: name,
		Contents: msg,
		TimeRecv: time.Now(),
	}
}


// Initializer to create a new QBot
func NewBot(path string) (*QBot, error) {
	// reads from the file
	file, err := ioutil.ReadFile(path)
	if nil != err {
		return &QBot{}, err
	}
	
	config := &Settings{}

	// json parse attempt #2
	err = json.Unmarshal(file, config)
	if err != nil {
		return &QBot{}, err
	}

	// do initialization here
	qb := QBot {
		Config: config,
		MaxSize: 100,
		queue: list.New(),
		MsgRate: time.Duration(2/3) * time.Millisecond,
		StartTime: time.Now(),
	}

	return &qb, nil

}


// Debug info to show what information was read from a key file
func (qb *QBot) Debug() {
	fmt.Printf("--- Bot Debug information ---\n")
	fmt.Printf("Name: %s\n", qb.Config.Name)
	fmt.Printf("Channel: %s\n", qb.Config.Channel)
	fmt.Printf("Server: %s\n", qb.Config.Server)
	fmt.Printf("Port: %s\n", qb.Config.Port)
	fmt.Printf("Password: hidden\n")
	fmt.Printf("--- Ending Debug ---\n")
}


// Tell the bot to connect to it's target server
func (qb *QBot) Connect() {
	var err error
	fmt.Printf("Connecting to %s ... \n", qb.Config.Server)
	qb.conn, err = net.Dial("tcp", qb.Config.Server+":"+qb.Config.Port)
	if err != nil {
		log.Fatal("Cannot connect to %s", qb.Config.Server)
		return
	}
	fmt.Printf("Connected to %s\n", qb.Config.Server)
	qb.StartTime = time.Now()
}


// Disconnect and close the TCP port
func (qb *QBot) Disconnect() {
	qb.conn.Close()
	upTime := time.Now().Sub(qb.StartTime).Seconds()
	fmt.Printf("Closed connection, live for %fs", upTime)
}


// Tell the bot to Join it's target channel
func (qb *QBot) JoinChannel() {
	fmt.Printf("Joining #%s ...\n", qb.Config.Channel)
	qb.conn.Write([]byte("PASS " + qb.Config.Password + "\r\n"))
	qb.conn.Write([]byte("NICK " + qb.Config.Name + "\r\n"))
	qb.conn.Write([]byte("JOIN #" + qb.Config.Channel + "\r\n"))
	fmt.Printf("Joined channel #%s\n", qb.Config.Channel)
}


// Read from the TCP port and process data on an infinite loop
func (qb *QBot) ReadPort() error {
	fmt.Printf("Watching #%s ...\n", qb.Config.Channel)

	tp := textproto.NewReader(bufio.NewReader(qb.conn))

	for {
		line, err := tp.ReadLine()
		if err != nil {
			qb.Disconnect()
			log.Fatal(err)
			return errors.New("HandleChatMessage: failed to read line from channel.")
		}

		// TODO: create a toggle parameter to display full IRC output?
		//fmt.Println(line)

		if line == "PING :tmi.twitch.tv" {
			fmt.Printf("Received PING, replying\n")
			_, err := qb.conn.Write([]byte("PONG :tmi.twitch.tv\r\n"))
			if err != nil {
				log.Fatal("Couldn't write PONG reply to server")
			}
			continue
		} else {
			matches := MsgRegex.FindStringSubmatch(line)
			if matches != nil {

				userName := matches[1]
				msgType := matches[2]

				switch msgType {
				case "PRIVMSG":
					msg := matches[3]
					fmt.Printf("[%s] %s\n", userName, msg)

					msgThing := NewMsg(userName, msg)
					err := qb.ProcessMsg(msgThing)
					if err != nil {
						log.Fatal(err)
					}

				default:
					// nothing of note, ignore
				}
			}
		}
		time.Sleep(qb.MsgRate)
	}
}

// When receiving text input, determine if any of it should dispatch
// to QBot commands
func (qb *QBot) ProcessMsg(msg Message) error {
	// here is where we will delegate commands based on messages

	// split the arguments up
	splits := strings.Split(msg.Contents, " ")
	var err error

	// TODO: fully implement command dispatch
	//fmt.Println("Split[0]:", splits[0])
	switch splits[0] {
	case "!test":
		err = qb.Say("Showing test data")
	case "!join":
		err = qb.Say("Joined the queue")
	case "!leave":
		err = qb.Say("Left the queue")
	case "!queue":
		err = qb.Say("showing queue")
	default:
		// nothing, no command found
	}

	// check if message is from owner, then check against owner-only funcs 

	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}


// TODO: finish implementation
func (qb *QBot) Push(p Player) error {
	if qb.queue.Len() >= qb.MaxSize {
		return errors.New("Push: queue at maximum limit")
	}

	qb.queue.PushBack(p)
	return nil
}


// TODO: finish implementation
func (qb *QBot) Pop(popsize int) error {
	qlen := qb.queue.Len()
	
	if qlen == 0 {
		return errors.New("Pop: queue is empty")
	}

	if qlen <= popsize {
		return errors.New("Pop: requested size greater than queue")
	}

	var msg string

	for i := 0; i < popsize; i++ {
		msg += "string "
		fmt.Println("Attempting to pop a player")
	}

	
	return nil
}

// Display/print the current queue to the chatroom
func (qb *QBot) ShowQueue() error {
	return nil
}


// Print a message out from the Bot
func (qb *QBot) Say(msg string) error {
	if msg == "" {
		return errors.New("Say: message was empty")
	}

	_, err := qb.conn.Write([]byte(fmt.Sprintf("PRIVMSG #%s :%s\r\n", qb.Config.Channel, msg)))
	if err != nil {
		return err
	}
	return nil
}


func (qb *QBot) Start() {
	for {
		qb.Connect()
		qb.JoinChannel()
		err := qb.ReadPort()
		if err != nil {
			time.Sleep(1000 * time.Millisecond)
			log.Fatal(err)
			fmt.Println("Starting up again...")
		} else {
			return
		}
		
	}
}

// quick timestamp function to properly format a date
func timeStamp() string {
	return time.Now().Format(ESTFormat)
}



// Main function to be run on program start
func main() {
	fmt.Println("=== QBot version 0.1 ===")
	fmt.Println("Parsing arguments")

	args := os.Args
	if len(args) != 2 {
		log.Fatal(errors.New("QBot.main(): No input file supplied"))
		return
	}

	bot, err := NewBot(args[1]) 
	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Debug()
	bot.Start()
	
}


// end qbot.go
