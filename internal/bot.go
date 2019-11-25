package internal

import (
	"fmt"
	"net"
	"net/textproto"
	"encoding/json"
	"strings"
	"regexp"
	"errors"
	"log"
	//"os"
	//"io"
	//"io/ioutil"
	"strconv"
	"bufio"
	"time"
	"container/list"
)



// Default key file to look for in the same directory
const DefaultKeyFile = "botdata.json"

// Message parsing and date stuff
const ESTFormat = "Jan 2 15:04:05 EST"
var MsgRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)
var CmdRegex *regexp.Regexp = regexp.MustCompile(`^!(\w+)\s?(\w+)?`)


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


// Initializer to create a new QBot
func NewBot(str *[]byte) (*QBot, error) {
	
	// define an empty settings struct
	config := &Settings{}

	// json parse attempt #2
	err := json.Unmarshal(*str, config)
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
					err := qb.ProcessMsg(&msgThing)
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
func (qb *QBot) ProcessMsg(msg *Message) error {
	// here is where we will delegate commands based on messages

	// split the arguments up
	splits := strings.Split(msg.Contents, " ")
	var err error

	// TODO: fully implement command dispatch
	//fmt.Println("Split[0]:", splits[0])
	switch splits[0] {
	case "!join":
		err = qb.JoinQueue(msg, splits[1:])
	case "!leave":
		err = qb.LeaveQueue(msg, splits[1:])
	case "!queue":
		err = qb.ShowQueue(msg, splits[1:])
	default:
		// nothing, no command found
	}
	
	// check if message is from owner, then check against owner-only funcs 
	if qb.Config.Name == msg.Name {
		switch splits[0] {
		case "!pop":
			err = qb.PopPlayers(msg, splits[1:])
		default:
			// nothing else
		}
	}

	return err
}


// Display/print the current queue to the chatroom
func (qb *QBot) ShowQueue(msg *Message, args []string) error {
	qlen := qb.queue.Len()
	node := qb.queue.Front()

	if node == nil {
		qb.Say("Queue is empty")
		return nil
	}

	out := "Queue: "
	var i int

	for i = 0; node != nil && i < 3; i++ {
		p := node.Value.(Player)
		out += p.Name

		node = node.Next()
		if node != nil {
			switch i {
			case 2:
				out += " ..."
			default:
				out += ", "
			}
		}
	}

	if i < qlen {
		out += fmt.Sprintf(" (%s more)", (qlen-i))
	}

	qb.Say(out)
	return nil
}


// Join the queue
func (qb *QBot) JoinQueue(msg *Message, args []string) error {
	fmt.Printf("Received request to join queue from %s\n", msg.Name)

	if qb.queue.Len() >= qb.MaxSize {
		return errors.New("Push: queue at maximum limit")
	}

	p := msg.ToPlayer()
	node := qb.queue.Front()

	for i := 0; node != nil; i++ {
		o := node.Value.(Player)
		if p.Name == o.Name {
			fmt.Printf("%s already in queue, skipping\n", p.Name)
			return nil
		}
		node = node.Next()
	}

	qb.queue.PushBack(msg.ToPlayer())
	qb.Say(fmt.Sprintf("%s has joined the queue", p.Name))
	return nil
}


// Leave the queue
func (qb *QBot) LeaveQueue(msg *Message, args []string) error {
	fmt.Printf("Received request to leave queue from %s\n", msg.Name)

	p := msg.ToPlayer()
	node := qb.queue.Front()

	// scan the entire queue to find our target player to remove
	for i := 0; node != nil; i++ {
		o := node.Value.(Player)

		if p.Name == o.Name {
			qb.Say(fmt.Sprintf("%s has left the queue", p.Name))
			qb.queue.Remove(node)
			return nil
		}
		node = node.Next()
	}

	if node == nil {
		qb.Say(fmt.Sprintf("@%s you are not in queue", p.Name))
	}
	return nil
}


// Pop players from the front of the queue
// TODO: add variable amount of players to pop
func (qb *QBot) PopPlayers(msg *Message, args []string) error {
	fmt.Printf("Received request to pop a player off the stack\n")

	qlen := qb.queue.Len()

	// check if queue is empty so we can skip doing any work
	if qlen == 0 {
		return errors.New("Pop: queue is empty")
	}

	// check if args were supplied via the command
	// i.e. !pop 5
	var popsize int
	var err error
	if len(args) > 0 {
		popsize, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Couldn't convert %s to an integer", args[0])
			popsize = 1
		}
	} else {
		popsize = 1
	}

	if popsize > qlen {
		fmt.Printf("Pop: requested size greater than queue, popping entire queue")
		popsize = qlen
	}

	node := qb.queue.Front()
	out := "Next player(s): "

	for i := 0; i < popsize; i++ {
		// proceed to the next node here
		c := node.Value.(Player)
		out += c.Name

		if i < (popsize-1) {
			out += ", "
		}

		// remove the node
		qb.queue.Remove(node)
		node = qb.queue.Front()
	}

	fmt.Print(out)
	qb.Say(out)

	return nil
}


// Print a message out from the Bot
func (qb *QBot) Say(msg string) error {
	if msg == "" {
		return errors.New("Say: message was empty")
	}

	if len(msg) > 500 {
		return errors.New("Say: message size exceeded byte limit of 500")
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




// end
