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
	"encoding/json"
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
	if index := linkMarkdownRegexShort.FindStringIndex(text); index != nil {
		submatch := linkMarkdownRegexShort.FindStringSubmatch(text)
		text = text[:index[0]] + submatch[1] + text[index[1]:]
	}

	// Remove links of the format "<http://google.com|google.com>"
	if index := linkMarkdownRegexLong.FindStringIndex(text); index != nil {
		submatch := linkMarkdownRegexLong.FindStringSubmatch(text)
		text = text[:index[0]] + submatch[1] + text[index[1]:]
	}

	return text
}

func buildSlackOptions(elements *templates.OutputElements) ([]slack.MsgOption, error) {
	options := []slack.MsgOption{
		slack.MsgOptionDisableMediaUnfurl(),
		slack.MsgOptionAsUser(false),
	}

	var blocks []slack.Block

	for _, e := range elements.Elements {
		switch t := e.(type) {
		case *templates.Divider:
			blocks = append(blocks, slack.NewDividerBlock())

		case *templates.Header:
			elements.Color = t.Color
			elements.Title = t.Title

			if elements.Title != "" {
				text := &templates.Text{
					Markdown: true,
					Text:     fmt.Sprintf("*%s*", t.Title),
				}
				textBlock, err := buildTextBlockObject(text)
				if err != nil {
					return nil, err
				}

				blocks = append(blocks, slack.NewSectionBlock(textBlock, nil, nil))
			}

		case *templates.Image:
			blocks = append(blocks, slack.NewImageBlock(t.URL, "alt-text", "", nil))

		case *templates.Section:
			var tbf []*slack.TextBlockObject
			var tbo *slack.TextBlockObject // TODO(mtitmus) There's currently no way for a user to set this.
			var tba *slack.Accessory

			if t.Text != nil {
				textBlock, err := buildTextBlockObject(t.Text)
				if err != nil {
					return nil, err
				}

				tbo = textBlock
			}

			for _, tf := range t.Fields {
				switch t := tf.(type) {
				case *templates.Text:
					if textBlock, err := buildTextBlockObject(t); err != nil {
						return nil, err
					} else {
						tbf = append(tbf, textBlock)
					}
				case *templates.Image:
					if tba == nil {
						tba = &slack.Accessory{}
					}
					tba.ImageElement = slack.NewImageBlockElement(t.URL, "alt-text")
				default:
					return nil, fmt.Errorf("%T elements are not supported inside a Section for Slack", e)
				}
			}

			blocks = append(blocks, slack.NewSectionBlock(tbo, tbf, tba))

		case *templates.Text:
			textBlock, err := buildTextBlockObject(t)
			if err != nil {
				return nil, err
			}

			blocks = append(blocks, slack.NewSectionBlock(textBlock, nil, nil))

		default:
			return nil, fmt.Errorf("%T elements are not yet supported by Gort for Slack", e)
		}
	}

	if elements.Color != "" {
		b, _ := json.MarshalIndent(elements, "", "  ")
		fmt.Println("A\n", string(b))
		b, _ = json.MarshalIndent(blocks, "", "  ")
		fmt.Println("A\n", string(b))

		attachment := slack.Attachment{
			Color:  elements.Color,
			Blocks: slack.Blocks{BlockSet: blocks},
		}

		options = append(options, slack.MsgOptionAttachments(attachment))
	} else {
		b, _ := json.MarshalIndent(elements, "", "  ")
		fmt.Println("B\n", string(b))
		b, _ = json.MarshalIndent(blocks, "", "  ")
		fmt.Println("B\n", string(b))
		options = append(options, slack.MsgOptionBlocks(blocks...))
	}

	return options, nil
}

func buildTextBlockObject(t *templates.Text) (*slack.TextBlockObject, error) {
	var textType string
	var emoji = t.Emoji

	if t.Markdown {
		textType = "mrkdwn"
		emoji = false
	} else {
		textType = "plain_text"
	}

	txt := t.Text
	if t.Monospace {
		txt = fmt.Sprintf("```%s```", txt)
	}

	tbo := slack.NewTextBlockObject(textType, txt, emoji, false)

	return tbo, tbo.Validate()
}
