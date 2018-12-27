package main

import (
	"fmt"

	"github.com/nlopes/slack"
)

func main() {
	api := slack.New("xoxb-318612539623-512030882149-I3JUlQSw30088hMU7fDQsScO")
	attachment := slack.Attachment{
		Pretext: "some pretext",
		Text:    "some text",
		// Uncomment the following part to send a field too
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: "a",
				Value: "no",
			},
		},
	}

	channelID, timestamp, err := api.PostMessage("C9BGF4YFL",
		slack.MsgOptionText("Some option text", false),
		slack.MsgOptionAttachments(attachment),
	)

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
}
