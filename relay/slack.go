package relay

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/clockworksoul/cog2/config"
	"github.com/clockworksoul/cog2/context"
	"github.com/nlopes/slack"
)

type SlackRelay struct {
	provider config.SlackProvider
	client   *slack.Client
	rtm      *slack.RTM

	_channels map[string]*slack.Channel
	_users    map[string]*slack.User
}

func NewSlackRelay(p config.SlackProvider) *SlackRelay {
	return &SlackRelay{provider: p}
}

func (s *SlackRelay) Connect() {
	log.Printf("Connecting to Slack provider %s...\n", s.provider.Name)

	go s.rtm.ManageConnection()
}

func (s *SlackRelay) Initialize() {
	s.client = slack.New(s.provider.SlackAPIToken)
	s.rtm = s.client.NewRTM()
}

func (s *SlackRelay) Listen() {
	for msg := range s.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			go s.OnConnected(ev)

		case *slack.MessageEvent:
			s.OnMessage(ev)

		case *slack.PresenceChangeEvent:
			log.Printf("Presence Change: %v\n", ev)

		case *slack.RTMError:
			s.OnError(ev)

		case *slack.InvalidAuthEvent:
			s.OnBadCredentials(ev)
			return

		default:
			// Ignore other events..
		}
	}
}

func (s *SlackRelay) OnBadCredentials(event *slack.InvalidAuthEvent) {
	log.Printf("Connection failed to %s: invalid credentials\n", s.provider.Name)
}

func (s *SlackRelay) OnConnected(event *slack.ConnectedEvent) {
	log.Printf(
		"Connection established to %s.slack.com. I am @%s!\n",
		event.Info.Team.Domain,
		event.Info.User.Name,
	)

	channels, err := s.rtm.GetChannels(true)
	if err != nil {
		log.Printf("Failed to get channels list for %s: %s", s.provider.Name, err.Error())
		return
	}

	for _, c := range channels {
		message := fmt.Sprintf("Cog2 version %s is online. Hello, %s!", context.CogVersion, c.Name)
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(message, c.ID))
	}
}

func (s *SlackRelay) OnError(event *slack.RTMError) {
	log.Println("Received error event: " + event.Error())
}

func (s *SlackRelay) OnMessage(event *slack.MessageEvent) {
	// If the message changed or deleted (or unknown), we ignore it.
	switch event.Msg.SubType {
	case "": // Base case. Do nothing.
	case "message_changed":
		fallthrough
	case "message_deleted":
		log.Printf("Message %s: ignoring", strings.TrimPrefix(event.Msg.SubType, "message_"))
		return
	default:
		log.Printf("Received unknown submessage type!")
		return
	}

	if event.Channel[0] == 'D' {
		s.OnDirectMessage(event)
	} else {
		s.OnChannelMessage(event)
	}
}

func (s *SlackRelay) OnChannelMessage(event *slack.MessageEvent) {
	channel, err := s.getChannelInfo(event.Channel)
	if err != nil {
		log.Printf("Could not find channel: " + err.Error())
		return
	}

	userinfo, err := s.getUserInfo(event.Msg.User)
	if err != nil {
		log.Printf("Could not find user: " + err.Error())
		return
	}

	log.Printf("Message from @%s in %s: %s\n",
		userinfo.Profile.DisplayNameNormalized,
		channel.Name,
		event.Msg.Text,
	)
}

func (s *SlackRelay) OnDirectMessage(event *slack.MessageEvent) {
	userinfo, err := s.getUserInfo(event.Msg.User)
	if err != nil {
		log.Printf("Could not find user: " + err.Error())
		return
	}

	log.Printf("Direct message from @%s: %s\n",
		userinfo.Profile.DisplayNameNormalized,
		event.Msg.Text,
	)
}

func (s *SlackRelay) getChannelInfo(id string) (*slack.Channel, error) {
	if id == "" {
		return nil, errors.New("Empty channel id")
	}

	if channel, ok := s._channels[id]; ok {
		return channel, nil
	}

	channel, err := s.rtm.GetChannelInfo(id)
	if channel == nil {
		return nil, errors.New("No such channel: " + id)
	} else if err != nil {
		s._channels[id] = channel
	}

	if channel == nil {
		err = fmt.Errorf("No such channel: %s", id)
	}

	return channel, err
}

func (s *SlackRelay) getUserInfo(id string) (*slack.User, error) {
	if id == "" {
		return nil, errors.New("Empty user id")
	}

	if user, ok := s._users[id]; ok {
		return user, nil
	}

	user, err := s.rtm.GetUserInfo(id)
	if user == nil {
		return nil, errors.New("No such user: " + id)
	} else if err != nil {
		s._users[id] = user
	}

	if user == nil {
		err = fmt.Errorf("No such user: %s", id)
	}

	return user, err
}
