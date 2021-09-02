package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var api = slack.New("TOKEN")

func main() {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	// Handle /command
	http.HandleFunc("/command", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("[INFO] Heard command")
		fmt.Println("[INFO] Start reading body")

		// Ready Body
		body, err := ioutil.ReadAll(r.Body)
		fmt.Println("[INFO] End reading body")
		if err != nil {
			fmt.Println("[ERROR] Unable to read body")
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify Secret
		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			fmt.Println("[ERROR] Unable to verify secret")
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Check for Server Error
		if _, err := sv.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			fmt.Println("[ERROR] Unauthorized")
			fmt.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Parse Event
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			fmt.Println("[ERROR] Error parsing event")
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// If URLVerification
		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				fmt.Println("[ERROR] Error unmarshaling json")
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}

		// If CallbackEvent
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
			}
		}
	})

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":8080", nil)
}
