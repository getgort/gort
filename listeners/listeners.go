package listeners

import (
	"fmt"
	"log"

	"github.com/clockworksoul/cog2/config"
	"github.com/nlopes/slack"
)

const (
	Version = "0.0.0"
)

type SlackListener struct {
	provider config.SlackProvider
	client   *slack.Client
	rtm      *slack.RTM

	_channels map[string]*slack.Channel
	_users    map[string]*slack.User
}

func NewSlackListener(p config.SlackProvider) *SlackListener {
	return &SlackListener{provider: p}
}

func (s *SlackListener) Connect() {
	log.Printf("Connecting to Slack provider %s...\n", s.provider.Name)

	go s.rtm.ManageConnection()
}

func (s *SlackListener) Initialize() {
	s.client = slack.New(s.provider.SlackAPIToken)
	s.rtm = s.client.NewRTM()
}

func (s *SlackListener) Listen() {
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

func (s *SlackListener) OnBadCredentials(event *slack.InvalidAuthEvent) {
	log.Printf("Connection failed to %s: invalid credentials\n", s.provider.Name)
}

func (s *SlackListener) OnConnected(event *slack.ConnectedEvent) {
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
		message := fmt.Sprintf("Cog2v%s is online. Hello, %s!", Version, c.Name)
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(message, c.ID))
	}
}

func (s *SlackListener) OnError(event *slack.RTMError) {
	log.Println("Received error event: " + event.Error())
}

func (s *SlackListener) OnMessage(event *slack.MessageEvent) {
	if event.Channel[0] == 'D' {
		s.OnDirectMessage(event)
	} else {
		s.OnChannelMessage(event)
	}
}

func (s *SlackListener) OnChannelMessage(event *slack.MessageEvent) {
	msg := event.Msg
	messageUpdated := false

	channel, err := s.getChannelInfo(event.Channel)
	if err != nil {
		log.Println("No channel found for ID " + event.Channel)
	}

	// If the message changed, respond to the update.
	// Is this the right thing to do?
	if msg.SubType == "message_changed" {
		msg = *event.SubMessage
		messageUpdated = true
	}

	userinfo, err := s.getUserInfo(msg.User)
	if err != nil {
		log.Println("No user found for ID " + msg.User)
	}

	if messageUpdated {
		log.Printf("Message from @%s in %s updated: %s\n",
			userinfo.Profile.DisplayNameNormalized,
			channel.Name,
			msg.Text,
		)
	} else {
		log.Printf("Message from @%s in %s: %s\n",
			userinfo.Profile.DisplayNameNormalized,
			channel.Name,
			msg.Text,
		)
	}
}

func (s *SlackListener) OnDirectMessage(event *slack.MessageEvent) {
	msg := event.Msg
	messageUpdated := false

	// If the message changed, respond to the update.
	// Is this the right thing to do?
	if msg.SubType == "message_changed" {
		msg = *event.SubMessage
		messageUpdated = true
	}

	userinfo, err := s.getUserInfo(msg.User)
	if err != nil {
		log.Println("No user found for ID " + msg.User)
	}

	if messageUpdated {
		log.Printf("Direct message from @%s updated: %s\n",
			userinfo.Profile.DisplayNameNormalized,
			msg.Text,
		)
	} else {
		log.Printf("Direct message from @%s: %s\n",
			userinfo.Profile.DisplayNameNormalized,
			msg.Text,
		)
	}
}

func (s *SlackListener) getChannelInfo(id string) (*slack.Channel, error) {
	if channel, ok := s._channels[id]; ok {
		return channel, nil
	} else {
		channel, err := s.rtm.GetChannelInfo(id)
		if err != nil {
			s._channels[id] = channel
		}

		if channel == nil {
			err = fmt.Errorf("No such channel: %s", id)
		}

		return channel, err
	}
}

func (s *SlackListener) getUserInfo(id string) (*slack.User, error) {
	if user, ok := s._users[id]; ok {
		return user, nil
	} else {
		user, err := s.rtm.GetUserInfo(id)
		if err != nil {
			s._users[id] = user
		}

		if user == nil {
			err = fmt.Errorf("No such user: %s", id)
		}

		return user, err
	}
}

func StartListening() {
	for _, sp := range config.GetSlackProviders() {
		listener := NewSlackListener(sp)

		listener.Initialize()
		listener.Connect()

		go listener.Listen()
	}
}
