package matterclient

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
	"github.com/mattermost/platform/model"
)

type Credentials struct {
	Login  string
	Team   string
	Pass   string
	Server string
}

type Message struct {
	Raw      *model.Message
	Post     *model.Post
	Team     string
	Channel  string
	Username string
	Text     string
}

type MMClient struct {
	*Credentials
	Client       *model.Client
	WsClient     *websocket.Conn
	Channels     *model.ChannelList
	MoreChannels *model.ChannelList
	User         *model.User
	Users        map[string]*model.User
	MessageChan  chan *Message
	//Team         *model.Team
}

func New(login, pass, team, server string) *MMClient {
	cred := &Credentials{Login: login, Pass: pass, Team: team, Server: server}
	mmclient := &MMClient{Credentials: cred, MessageChan: make(chan *Message, 100)}
	return mmclient
}

func (m *MMClient) Login() error {
	b := &backoff.Backoff{
		Min:    time.Second,
		Max:    5 * time.Minute,
		Jitter: true,
	}
	// login to mattermost
	m.Client = model.NewClient("https://" + m.Credentials.Server)
	var myinfo *model.Result
	var appErr *model.AppError
	for {
		log.Println("retrying login", m.Credentials.Team, m.Credentials.Login, m.Credentials.Server)
		myinfo, appErr = m.Client.LoginByEmail(m.Credentials.Team, m.Credentials.Login, m.Credentials.Pass)
		if appErr != nil {
			d := b.Duration()
			if !strings.Contains(appErr.DetailedError, "connection refused") &&
				!strings.Contains(appErr.DetailedError, "invalid character") {
				return errors.New(appErr.Message)
			}
			log.Printf("LOGIN: %s, reconnecting in %s", appErr, d)
			time.Sleep(d)
			continue
		}
		break
	}
	// reset timer
	b.Reset()
	m.User = myinfo.Data.(*model.User)
	/*
		myinfo, _ = MmClient.GetMyTeam("")
		u.MmTeam = myinfo.Data.(*model.Team)
	*/

	// setup websocket connection
	wsurl := "wss://" + m.Credentials.Server + "/api/v1/websocket"
	header := http.Header{}
	header.Set(model.HEADER_AUTH, "BEARER "+m.Client.AuthToken)

	var WsClient *websocket.Conn
	var err error
	for {
		WsClient, _, err = websocket.DefaultDialer.Dial(wsurl, header)
		if err != nil {
			d := b.Duration()
			log.Printf("WSS: %s, reconnecting in %s", err, d)
			time.Sleep(d)
			continue
		}
		break
	}
	b.Reset()

	m.WsClient = WsClient

	// populating users
	m.updateUsers()

	// populating channels
	m.updateChannels()

	return nil
}

func (m *MMClient) WsReceiver() {
	var rmsg model.Message
	for {
		if err := m.WsClient.ReadJSON(&rmsg); err != nil {
			log.Println("error:", err)
			// reconnect
			m.Login()
		}
		//log.Printf("WsReceiver: %#v", rmsg)
		msg := &Message{Raw: &rmsg, Team: m.Team}
		m.parseMessage(msg)
		m.MessageChan <- msg
	}

}

func (m *MMClient) parseMessage(rmsg *Message) {
	switch rmsg.Raw.Action {
	case model.ACTION_POSTED:
		m.parseActionPost(rmsg)
		/*
			case model.ACTION_USER_REMOVED:
				m.handleWsActionUserRemoved(&rmsg)
			case model.ACTION_USER_ADDED:
				m.handleWsActionUserAdded(&rmsg)
		*/
	}
}

func (m *MMClient) parseActionPost(rmsg *Message) {
	data := model.PostFromJson(strings.NewReader(rmsg.Raw.Props["post"]))
	//	log.Println("receiving userid", data.UserId)
	// we don't have the user, refresh the userlist
	if m.Users[data.UserId] == nil {
		m.updateUsers()
	}
	rmsg.Username = m.Users[data.UserId].Username
	rmsg.Channel = m.getChannelName(data.ChannelId)
	// direct message
	if strings.Contains(rmsg.Channel, "__") {
		//log.Println("direct message")
		rcvusers := strings.Split(rmsg.Channel, "__")
		if rcvusers[0] != m.User.Id {
			rmsg.Channel = m.Users[rcvusers[0]].Username
		} else {
			rmsg.Channel = m.Users[rcvusers[1]].Username
		}
	}
	rmsg.Text = data.Message
	rmsg.Post = data
	return
}

func (m *MMClient) updateUsers() error {
	mmusers, _ := m.Client.GetProfiles(m.User.TeamId, "")
	m.Users = mmusers.Data.(map[string]*model.User)
	return nil
}

func (m *MMClient) updateChannels() error {
	mmchannels, _ := m.Client.GetChannels("")
	m.Channels = mmchannels.Data.(*model.ChannelList)
	mmchannels, _ = m.Client.GetMoreChannels("")
	m.MoreChannels = mmchannels.Data.(*model.ChannelList)
	return nil
}

func (m *MMClient) getChannelName(id string) string {
	for _, channel := range append(m.Channels.Channels, m.MoreChannels.Channels...) {
		if channel.Id == id {
			return channel.Name
		}
	}
	// not found? could be a new direct message from mattermost. Try to update and check again
	m.updateChannels()
	for _, channel := range append(m.Channels.Channels, m.MoreChannels.Channels...) {
		if channel.Id == id {
			return channel.Name
		}
	}
	return ""
}

func (m *MMClient) getChannelId(name string) string {
	for _, channel := range append(m.Channels.Channels, m.MoreChannels.Channels...) {
		if channel.Name == name {
			return channel.Id
		}
	}
	return ""
}

func (m *MMClient) PostMessage(channel string, text string) {
	post := &model.Post{ChannelId: m.getChannelId(channel), Message: text}
	m.Client.CreatePost(post)
}
