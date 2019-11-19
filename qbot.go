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
	"io"
	"io/ioutil"
	"bufio"
	"time"
	//"container/list"
)

const ESTFormat = "Jan 2 15:04:05 EST"
var MsgRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)

var CmdRegex *regexp.Regexp = regexp.MustCompile(`^!(\w+)\s?(\w+)?`)


// A very basic Bot interface for any chat bots
type Bot interface {
	Connect()
	Disconnect()
	JoinChannel()
	ReadPort() error
	ProcessMsg(msg Message) error
	ReadCredentials() error
	Say(msg string)
	Start()
}

type OAuthCred struct {
	Password string `json:"password,omitempty"`
	ClientID string `json:"client_id,omitempty"`
}

// Our QBot, who stores a queue of players
type QBot struct {
	conn           net.Conn
	Channel        string
	Credentials    *OAuthCred
	MsgRate        time.Duration
	Name           string
	Port           string
	PrivatePath    string
	Server         string
	StartTime      time.Time
}

type Message struct {
	Name      string
	Contents  string
	TimeRecv  time.Time
}

type Player struct {
	Name string
	TimeJoined string
}

func (qb *QBot) Connect() {
	var err error
	fmt.Println("Connecting to %s ... ", qb.Server)
	qb.conn, err = net.Dial("tcp", qb.Server+":"+qb.Port)
	if err != nil {
		fmt.Println("Cannot connect to %s", qb.Server)
		return
	}
	fmt.Println("Connected to %s", qb.Server)
	qb.StartTime = time.Now()
}

func (qb *QBot) Disconnect() {
	qb.conn.Close()
	upTime := time.Now().Sub(qb.StartTime).Seconds()
	fmt.Println("Closed connection, live for %fs", upTime)
}

func (qb *QBot) JoinChannel() {
	fmt.Println("Joining #%s ...", qb.Channel)
	qb.conn.Write([]byte("PASS " + qb.Credentials.Password + "\r\n"))
	qb.conn.Write([]byte("NICK " + qb.Name + "\r\n"))
	qb.conn.Write([]byte("JOIN #" + qb.Channel + "\r\n"))
	fmt.Println("Joined channel #%s", qb.Channel)
}

func (qb *QBot) ReadCredentials() error {
	credFile, err := ioutil.ReadFile(qb.PrivatePath)
	if err != nil {
		return err
	}

	qb.Credentials = &OAuthCred{}

	dec := json.NewDecoder(strings.NewReader(string(credFile)))
	if err = dec.Decode(qb.Credentials); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (qb *QBot) ReadPort() error {
	fmt.Println("Watching #%s ...", qb.Channel)

	tp := textproto.NewReader(bufio.NewReader(qb.conn))

	for {
		line, err := tp.ReadLine()
		if err != nil {
			qb.Disconnect()
			log.Fatal(err)
			return errors.New("HandleChatMessage: failed to read line from channel.")
		}

		if line == "PING :tmi.twitch.tv" {
			qb.conn.Write([]byte("PONG :tmi.twitch.tv\r\n"))
		} else {
			matches := MsgRegex.FindStringSubmatch(line)
			if matches != nil {

				userName := matches[1]
				msgType := matches[2]

				switch msgType {
				case "PRIVMSG":
					msg := matches[3]
					fmt.Println("[%s] %s", userName, msg)

					msgThing := Message{ 
						Name: userName,
						Contents: msg,
						TimeRecv: time.Now(),
					}

					err := qb.ProcessMsg(msgThing)
					if err != nil {
						log.Fatal(err)
					}

				default:
					// nothing
				}
			}
		}
		time.Sleep(qb.MsgRate)
	}
}

func (qb *QBot) ProcessMsg(msg Message) error {


	return nil
}

func (qb *QBot) Say(msg string) error {
	if msg == "" {
		return errors.New("Say: message was empty")
	}

	_, err := qb.conn.Write([]byte(fmt.Sprintf("PRIVMSG #%s %s\r\n", qb.Channel, msg)))
	if err != nil {
		return err
	}
	return nil
}


func (qb *QBot) Start() {
	err := qb.ReadCredentials()
	if err != nil {
		log.Fatal(err)
		return
	}

	for {
		qb.Connect()
		qb.JoinChannel()
		err = qb.ReadPort()
		if err != nil {
			time.Sleep(1000 * time.Millisecond)
			log.Fatal(err)
			fmt.Println("Starting up again...")
		} else {
			return
		}
		
	}
}

func timeStamp() string {
	return time.Now().Format(ESTFormat)
}



func main() {
	fmt.Println("Hello!")
}
