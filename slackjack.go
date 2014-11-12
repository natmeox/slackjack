package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"net/http"
	"strings"
)

var Config struct {
	Debug        bool
	SlackToken  string
	SlackTeam     string
	WebAddress   string
}

type SlackMessage struct {
	ChannelName string
	UserName string
	UserId string
	Text string
	Trigger string
	Timestamp string
}

type SlackResponse struct {
	Text string `json:"text"`
}

func JackRespond(msg *SlackMessage) (string, error) {

	switch strings.ToLower(msg.Text) {
	case "hit me":
		return JackHit(msg)
	case "deal":
		return JackDeal(msg)
	}

	return fmt.Sprintf("Not sure what you mean by “%s”, friend.", msg.Text), nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.json", "path to configuration file")

	flag.Parse()

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Println(err)
		return
	}
	dec := json.NewDecoder(configFile)
	err = dec.Decode(&Config)
	if err != nil {
		log.Println("Error decoding configuration:", err)
		return
	}

	http.HandleFunc("/jack/", func (w http.ResponseWriter, req *http.Request) {
		trigger := req.PostFormValue("trigger_word")
		fullText := req.PostFormValue("text")
		text := strings.TrimSpace(strings.TrimPrefix(fullText, trigger))

		msg := &SlackMessage{
			ChannelName: req.PostFormValue("channel_name"),
			UserName: req.PostFormValue("user_name"),
			UserId: req.PostFormValue("user_id"),
			Timestamp: req.PostFormValue("timestamp"),
			Trigger: trigger,
			Text: text,
		}

		ret, err := JackRespond(msg)
		if err != nil {
			ret = fmt.Sprintf("Oops: %s", err.Error())
		}

		response := &SlackResponse{
			Text: ret,
		}
		responseText, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "text/json")
		w.Write([]byte(responseText))
	})

	InitCards()

	http.ListenAndServe(Config.WebAddress, nil)
}
