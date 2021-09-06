package slack

import (
	"regexp"

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/data"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

var (
	linkMarkdownRegexShort = regexp.MustCompile(`\<([^|:]*:[^|]*)\>`)
	linkMarkdownRegexLong  = regexp.MustCompile(`\<[^|:]*:[^|]*\|([^|]*)\>`)
)

// NewAdapter will construct a SlackAdapter instance for a given provider configuration.
func NewAdapter(provider data.SlackProvider) adapter.Adapter {
	if provider.APIToken != "" {
		log.Warn("Classic Slack apps are deprecated, please upgrade to a Socket mode app.")

		client := slack.New(provider.APIToken)
		rtm := client.NewRTM()

		return ClassicAdapter{
			client:   client,
			provider: provider,
			rtm:      rtm,
		}
	}

	client := slack.New(
		provider.BotToken,
		slack.OptionAppLevelToken(provider.AppToken),
	)

	socketClient := socketmode.New(client)

	return &SocketModeAdapter{
		provider:     provider,
		client:       client,
		socketClient: socketClient,
	}
}

// ScrubMarkdown removes unnecessary/undesirable Slack markdown (of links, of
// example) from text received from Slack.
func ScrubMarkdown(text string) string {
	// Remove links of the format "<https://google.com>"
	//
	if index := linkMarkdownRegexShort.FindStringIndex(text); index != nil {
		submatch := linkMarkdownRegexShort.FindStringSubmatch(text)
		text = text[:index[0]] + submatch[1] + text[index[1]:]
	}

	// Remove links of the format "<http://google.com|google.com>"
	//
	if index := linkMarkdownRegexLong.FindStringIndex(text); index != nil {
		submatch := linkMarkdownRegexLong.FindStringSubmatch(text)
		text = text[:index[0]] + submatch[1] + text[index[1]:]
	}

	return text
}
