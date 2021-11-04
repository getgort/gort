/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package slack

import (
	"fmt"
	"regexp"

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/templates"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

var (
	linkMarkdownRegexShort = regexp.MustCompile(`\<([^|:]*:[^|]*)\>`)
	linkMarkdownRegexLong  = regexp.MustCompile(`\<[^|:]*:[^|]*\|([^|]*)\>`)
)

// const DefaultCommandTemplate = "```{{ .Response.Out }}```"

// const DefaultCommandErrorTemplate = "The pipeline failed planning the invocation:\n" +
// 	"```{{ .Request.Bundle.Name }}:{{ .Request.Command.Name }} {{ .Request.Parameters }}```\n" +
// 	"The specific error was:\n" +
// 	"```{{ .Response.Out }}```"

// const DefaultMessageTemplate = "{{ .Response.Out }}"

// NewAdapter will construct a SlackAdapter instance for a given provider configuration.
func NewAdapter(provider data.SlackProvider) adapter.Adapter {
	if provider.APIToken != "" {
		log.Warn("Classic Slack apps are deprecated, please upgrade to a Socket mode app.")

		client := slack.New(provider.APIToken)
		rtm := client.NewRTM()

		return &ClassicAdapter{
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

func buildSlackOptions(elements templates.OutputElements) ([]slack.MsgOption, error) {

	// slack.MsgOptionAttachments(
	// 	slack.Attachment{
	// 		// Title: "This is a title",
	// 		// Text:       "text",
	// 		Color: "#FF0000",
	// 		// MarkdownIn: []string{"text"},
	// 		Blocks: slack.Blocks{
	// 			BlockSet: []slack.Block{
	// 				slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "Hello! :wave: Enjoy your cat picture!", false, true), nil, nil),
	// 				slack.NewImageBlock(
	// 					img, "A kitty!", "",
	// 					slack.NewTextBlockObject("plain_text", "Please enjoy this cat picture.", false, false),
	// 				),
	// 			},
	// 		},
	// 	},
	// ),

	options := []slack.MsgOption{
		slack.MsgOptionDisableMediaUnfurl(),
		slack.MsgOptionAsUser(false),
	}

	attachment := slack.Attachment{
		Title: elements.Title,
		Color: elements.Color,
	}

	var blocks []slack.Block

	for _, e := range elements.Elements {
		switch t := e.(type) {
		case *templates.Divider:
			blocks = append(blocks, slack.NewDividerBlock())

		case *templates.Image:
			blocks = append(blocks, slack.NewImageBlock(t.URL, "alt-text", "", nil))

		case *templates.Section:
			var tbf []*slack.TextBlockObject
			var tbo *slack.TextBlockObject // TODO(mtitmus) There's currently no way for a user to set this.
			var tba *slack.Accessory = &slack.Accessory{}

			if t.Text != nil {
				tbo = buildTextBlockObject(t.Text)
			}

			for _, tf := range t.Fields {
				switch t := tf.(type) {
				case *templates.Text:
					tbf = append(tbf, buildTextBlockObject(t))
				case *templates.Image:
					tba.ImageElement = slack.NewImageBlockElement(t.URL, "alt-text")
				default:
					return nil, fmt.Errorf("%T elements are not supported inside a Section for Slack", e)
				}
			}

			blocks = append(blocks, slack.NewSectionBlock(tbo, tbf, tba))

		case *templates.Text:
			blocks = append(blocks, slack.NewSectionBlock(
				buildTextBlockObject(t), nil, nil,
			))

		default:
			return nil, fmt.Errorf("%T elements are not yet supported for Slack", e)
		}
	}

	if blocks != nil {
		attachment.Blocks = slack.Blocks{BlockSet: blocks}
	}

	options = append(options, slack.MsgOptionAttachments(attachment))

	return options, nil
}

func buildTextBlockObject(t *templates.Text) *slack.TextBlockObject {
	textType := "mrkdwn"
	if !t.Markdown {
		textType = "plain_text"
	}

	txt := t.Text
	if t.Monospace {
		txt = fmt.Sprintf("```%s```", txt)
	}

	return slack.NewTextBlockObject(textType, txt, t.Emoji, t.Markdown)
}
