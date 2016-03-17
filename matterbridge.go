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
	i      *irc.Connection
	m      *matterclient.MMClient
	ircMap map[string]string
	mmMap  map[string]string
	*Config
}

func NewBridge(name string, config *Config) *Bridge {
	b := &Bridge{}
	b.Config = config
	b.ircMap = make(map[string]string)
	b.mmMap = make(map[string]string)
	if len(b.Config.Channel) > 0 {
		for _, val := range b.Config.Channel {
			b.ircMap[val.IRC] = val.Mattermost
			b.mmMap[val.Mattermost] = val.IRC
		}
	}
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
	err := i.Connect(b.Config.IRC.Server + ":" + strconv.Itoa(b.Config.IRC.Port))
	if err != nil {
		log.Println(err)
	}
	time.Sleep(time.Second * 10)
	log.Println("Joining", b.Config.IRC.Channel, "as", b.Config.IRC.Nick)
	i.Join(b.Config.IRC.Channel)
	for _, val := range b.Config.Channel {
		log.Println("Joining", val.IRC, "as", b.Config.IRC.Nick)
		i.Join(val.IRC)
	}
	i.AddCallback("PRIVMSG", b.handlePrivMsg)
	i.AddCallback("CTCP_ACTION", b.handlePrivMsg)
	if b.Config.Mattermost.ShowJoinPart {
		i.AddCallback("JOIN", b.handleJoinPart)
		i.AddCallback("PART", b.handleJoinPart)
	}
	//	i.AddCallback("353", b.handleOther)
	return i
}

func (b *Bridge) handlePrivMsg(event *irc.Event) {
	msg := ""
	if event.Code == "CTCP_ACTION" {
		msg = event.Nick + " "
	}
	msg += event.Message()
	b.Send("irc-"+event.Nick, msg, b.getMMChannel(event.Arguments[0]))
}

func (b *Bridge) handleJoinPart(event *irc.Event) {
	b.Send(b.Config.IRC.Nick, "irc-"+event.Nick+" "+strings.ToLower(event.Code)+"s "+event.Message(), b.getMMChannel(event.Arguments[0]))
}

func (b *Bridge) handleOther(event *irc.Event) {
	switch event.Code {
	case "353":
		b.Send(b.Config.IRC.Nick, event.Message()+" currently on IRC", b.getMMChannel(event.Arguments[0]))
	}
}

func (b *Bridge) Send(nick string, message string, channel string) error {
	return b.SendType(nick, message, channel, "")
}

func (b *Bridge) SendType(nick string, message string, channel string, mtype string) error {
	b.m.PostMessage(channel, message)
	return nil
}

func (b *Bridge) handleMatter() {
	for message := range b.m.MessageChan {
		if message.Raw.Action == "posted" {
			cmd := strings.Fields(message.Text)[0]
			switch cmd {
			case "!users":
				log.Println("received !users from", message.Username)
				b.i.SendRaw("NAMES " + b.getIRCChannel(message.Channel))
			case "!gif":
				message.Text = b.giphyRandom(strings.Fields(strings.Replace(message.Text, "!gif ", "", 1)))
				b.Send(b.Config.IRC.Nick, message.Text, b.getIRCChannel(message.Channel))
			}
			texts := strings.Split(message.Text, "\n")
			for _, text := range texts {
				b.i.Privmsg(b.getIRCChannel(message.Channel), message.Username+": "+text)
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

func (b *Bridge) getMMChannel(ircChannel string) string {
	mmchannel, ok := b.ircMap[ircChannel]
	if !ok {
		mmchannel = b.Config.Mattermost.Channel
	}
	return mmchannel
}

func (b *Bridge) getIRCChannel(mmChannel string) string {
	ircchannel, ok := b.mmMap[mmChannel]
	if !ok {
		ircchannel = b.Config.IRC.Channel
	}
	return ircchannel
}

func main() {
	flagConfig := flag.String("conf", "matterbridge.conf", "config file")
	flag.Parse()
	NewBridge("matterbot", NewConfig(*flagConfig))
	select {}
}
