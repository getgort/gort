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
	"context"
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
// TODO(mtitmus) Can this be replaced by using Slack's "verbatim text" option?
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

// Send sends the contents of a response envelope to a
// specified channel. If channelID is empty the value of
// envelope.Request.ChannelID will be used.
func Send(ctx context.Context, client *slack.Client, a adapter.Adapter, channelID string, elements templates.OutputElements) error {
	e := log.WithContext(ctx)

	options, err := buildSlackOptions(&elements)
	if err != nil {
		e.WithError(err).Error("failed to build Slack options")
		if err := a.SendError(ctx, channelID, err); err != nil {
			e.WithError(err).Error("break-glass send error failure!")
		}
		return err
	}

	_, _, err = client.PostMessage(channelID, options...)
	if err != nil {
		e.WithError(err).Error("failed to post Slack message")
		if err := a.SendError(ctx, channelID, err); err != nil {
			e.WithError(err).Error("break-glass send error failure!")
		}
		return err
	}

	return nil
}

func SendError(ctx context.Context, client *slack.Client, channelID string, err error) error {
	_, _, e := client.PostMessage(
		channelID,
		slack.MsgOptionAttachments(
			slack.Attachment{
				Title:      "Error",
				Text:       err.Error(),
				Color:      "#FF0000",
				MarkdownIn: []string{"text"},
				ThumbURL:   "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/120/apple/285/fire_1f525.png",
			},
		),
		slack.MsgOptionDisableMediaUnfurl(),
		slack.MsgOptionDisableMarkdown(),
		slack.MsgOptionAsUser(false),
	)

	return e
}

// buildSlackOptions accepts a templates.OutputElements value produced by
// templates.EncodeElements and produces a roughly equivalent []slack.MsgOption
// value. It's used directly by the two adapter implementations.
func buildSlackOptions(elements *templates.OutputElements) ([]slack.MsgOption, error) {
	options := []slack.MsgOption{
		slack.MsgOptionDisableMediaUnfurl(),
		slack.MsgOptionAsUser(false),
	}

	var blocks []slack.Block
	var headerBlock *slack.SectionBlock

	for _, e := range elements.Elements {
		switch t := e.(type) {
		case *templates.Divider:
			blocks = append(blocks, slack.NewDividerBlock())

		case *templates.Header:
			elements.Color = t.Color
			elements.Title = t.Title

			if t.Title != "" {
				text := &templates.Text{Markdown: true, Text: fmt.Sprintf("*%s*", t.Title)}

				if textBlock, err := buildTextBlockObject(text); err != nil {
					return nil, err
				} else {
					headerBlock = slack.NewSectionBlock(textBlock, nil, nil)
				}
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

	// Slack attachments are funny. If you try to use a default one you get
	// an error. Also if you try to set a title and use blocks you get an
	// error. Therefore we only use one if we have

	// If there's no color set, we just use blocks and no attachment options.
	if elements.Color == "" {
		if headerBlock != nil {
			blocks = append([]slack.Block{headerBlock}, blocks...)
		}
		options = append(options, slack.MsgOptionBlocks(blocks...))

		return options, nil
	}

	// If we have one block, and it's a text block, we can use a title!
	// This isn't TECHNICALLY necessary, but it produces a cleaner output.
	if sb, ok := blocks[0].(*slack.SectionBlock); ok && sb.Text != nil && len(blocks) == 1 {
		attachment := slack.Attachment{
			Color: elements.Color,
			Title: elements.Title,
			Text:  sb.Text.Text,
		}
		options = append(options, slack.MsgOptionAttachments(attachment))

		return options, nil
	}

	// We CAN use an attachment to set a message color (without a title) and
	// still use blocks. Let's do that.

	if headerBlock != nil {
		blocks = append([]slack.Block{headerBlock}, blocks...)
	}
	attachment := slack.Attachment{
		Color:  elements.Color,
		Blocks: slack.Blocks{BlockSet: blocks},
	}
	options = append(options, slack.MsgOptionAttachments(attachment))

	return options, nil
}

// buildTextBlockObject accepts a templates.Text value, does some basic error
// correction to satisty the very tempermental Slack API, and returns an
// equivalent slack.TextBlockObject. It produces an error if the resulting
// TextBlockObject is not valid (according to TextBlockObject.Validate())
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
