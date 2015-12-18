package main

import (
	"crypto/tls"
	"flag"
	"github.com/42wim/matterbridge-plus/matterclient"
	"github.com/peterhellberg/giphy"
	"github.com/thoj/go-ircevent"
	"log"
	"strconv"
	"strings"
	"time"
)

type Bridge struct {
	i *irc.Connection
	m *matterclient.MMClient
	*Config
}

func NewBridge(name string, config *Config) *Bridge {
	b := &Bridge{}
	b.Config = config
	b.m = matterclient.New(b.Config.Mattermost.Login, b.Config.Mattermost.Password,
		b.Config.Mattermost.Team, b.Config.Mattermost.Server)
	err := b.m.Login()
	if err != nil {
		log.Fatal("can not connect", err)
	}
	go b.m.WsReceiver()

	b.i = b.createIRC(name)
	go b.handleMatter()
	return b
}

func (b *Bridge) createIRC(name string) *irc.Connection {
	i := irc.IRC(b.Config.IRC.Nick, b.Config.IRC.Nick)
	i.UseTLS = b.Config.IRC.UseTLS
	i.TLSConfig = &tls.Config{InsecureSkipVerify: b.Config.IRC.SkipTLSVerify}
	i.Connect(b.Config.IRC.Server + ":" + strconv.Itoa(b.Config.IRC.Port))
	time.Sleep(time.Second)
	log.Println("Joining", b.Config.IRC.Channel, "as", b.Config.IRC.Nick)
	i.Join(b.Config.IRC.Channel)
	i.AddCallback("PRIVMSG", b.handlePrivMsg)
	i.AddCallback("CTCP_ACTION", b.handlePrivMsg)
	if b.Config.Mattermost.ShowJoinPart {
		i.AddCallback("JOIN", b.handleJoinPart)
		i.AddCallback("PART", b.handleJoinPart)
	}
	i.AddCallback("353", b.handleOther)
	return i
}

func (b *Bridge) handlePrivMsg(event *irc.Event) {
	msg := ""
	if event.Code == "CTCP_ACTION" {
		msg = event.Nick + " "
	}
	msg += event.Message()
	b.Send("irc-"+event.Nick, msg)
}

func (b *Bridge) handleJoinPart(event *irc.Event) {
	b.SendType(b.Config.IRC.Nick, "irc-"+event.Nick+" "+strings.ToLower(event.Code)+"s "+event.Message(), "join_leave")
}

func (b *Bridge) handleOther(event *irc.Event) {
	switch event.Code {
	case "353":
		b.Send(b.Config.IRC.Nick, event.Message()+" currently on IRC")
	}
}

func (b *Bridge) Send(nick string, message string) error {
	return b.SendType(nick, message, "")
}

func (b *Bridge) SendType(nick string, message string, mtype string) error {
	b.m.PostMessage(b.Config.Mattermost.Channel, message)
	return nil
}

func (b *Bridge) handleMatter() {
	for message := range b.m.MessageChan {
		if message.Raw.Action == "posted" {
			if message.Channel == b.Config.Mattermost.Channel {
				cmd := strings.Fields(message.Text)[0]
				switch cmd {
				case "!users":
					log.Println("received !users from", message.User)
					b.i.SendRaw("NAMES " + b.Config.IRC.Channel)
				case "!gif":
					message.Text = b.giphyRandom(strings.Fields(strings.Replace(message.Text, "!gif ", "", 1)))
					b.Send(b.Config.IRC.Nick, message.Text)
				}
				texts := strings.Split(message.Text, "\n")
				for _, text := range texts {
					b.i.Privmsg(b.Config.IRC.Channel, message.User+": "+text)
				}
			}
		}
	}
}

func (b *Bridge) giphyRandom(query []string) string {
	g := giphy.DefaultClient
	if b.Config.General.GiphyAPIKey != "" {
		g.APIKey = b.Config.General.GiphyAPIKey
	}
	res, err := g.Random(query)
	if err != nil {
		return "error"
	}
	return res.Data.FixedHeightDownsampledURL
}

func main() {
	flagConfig := flag.String("conf", "matterbridge.conf", "config file")
	flag.Parse()
	NewBridge("matterbot", NewConfig(*flagConfig))
	select {}
}
