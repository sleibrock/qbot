package main


import (
	"os"
	"log"
	"fmt"
	"bufio"
	"io/ioutil"
	"encoding/json"

	"github.com/sleibrock/qbot/internal"
)


// Main function to be run on program start
func main() {
	fmt.Println("=== QBot version 0.1 ===")

	keyfile, err := ioutil.ReadFile(internal.DefaultKeyFile)
	if err != nil {
		// when no keyfile exists, create a default settings struct
		// then re-export the default struct to a file
		defsets := internal.DefaultSettings()
		js, err := json.MarshalIndent(defsets, "", "    ")
		if err != nil {
			log.Fatal(err)
		}

		// attempt to write to file
		err = ioutil.WriteFile(internal.DefaultKeyFile, js, 0644)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("You do not have a botdata.json file")
		fmt.Println("Please edit your botdata.json file with your")
		fmt.Println("Twitch settings and generate an OAuth password")
		fmt.Println("by connecting your Twitch account to ")
		fmt.Println("")
		fmt.Println("https://twitchapps.com/tmi/")
		fmt.Println("")
		fmt.Println("Once you connect, you will be given a password")
		fmt.Println("which you can put into the 'password' field of")
		fmt.Println("the JSON file.")
		fmt.Println("")
		fmt.Println("[Press any key to exit]")

		// eat a single line to quit
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		return
	}

	// attempt to load the keyfile into the bot
	bot, err := internal.NewBot(&keyfile) 
	if err != nil {
		log.Fatal(err)
		return
	}

	// turn this line off when it's no longer needed
	bot.Debug()

	// start the entire bot process
	bot.Start()
}


// end qbot.go
