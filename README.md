# matterbridge-plus

Simple bridge between mattermost and IRC. (Uses the mattermost API instead of webhooks)

* Relays public channel messages between mattermost and IRC.
* Supports multiple mattermost and irc channels.
* Matterbridge-plus also works with private groups on your mattermost.

## Requirements:
* [Mattermost] (https://github.com/mattermost/platform/) 2.1.0 (stable, not a dev build)
* A dedicated user(bot) on your mattermost instance.

There is also a version with webhooks that doesn't need a dedicated user. See [matterbridge] (https://github.com/42wim/matterbridge/)   

If you want to test with mattermost development builds, you also need to use the develop branch of matterbridge-plus.

## binaries
Binaries can be found [here] (https://github.com/42wim/matterbridge-plus/releases/tag/v0.1)

## building
Go 1.6 is required (or go1.5 with GO15VENDOREXPERIMENT=1)  
Make sure you have [Go](https://golang.org/doc/install) properly installed, including setting up your [GOPATH] (https://golang.org/doc/code.html#GOPATH)

```
cd $GOPATH
go get github.com/42wim/matterbridge-plus
```

You should now have matterbridge binary in the bin directory:

```
$ ls bin/
matterbridge-plus
```

## running
1) Copy the matterbridge.conf.sample to matterbridge.conf in the same directory as the matterbridge binary.  
2) Edit matterbridge.conf with the settings for your environment. See below for more config information.  
3) Now you can run matterbridge-plus.

```
Usage of matterbridge-plus:
  -conf="matterbridge.conf": config file
```

Matterbridge will:
* connect to specified irc server and channel.
* send messages from mattermost to irc and vice versa.

If you set PrefixMessagesWithNick to true, each message from IRC to Mattermost
will by default be prefixed by "irc-" + nick. You can, however, modify how the
messages appear, by setting (and modifying) RemoteNickFormat.
IRC.RemoteNickFormat defines how Mattermost nicks appear on IRC, and
Mattermost.RemoteNickFormat defines how IRC users appear on Mattermost. The
string "{NICK}" (case sensitive) will be replaced by the actual nick / username.

## config
### matterbridge-plus
matterbridge-plus looks for matterbridge.conf in current directory.

Look at matterbridge.conf.sample for an example

```
[IRC]
server="irc.freenode.net"
port=6667
UseTLS=false
SkipTLSVerify=true
nick="matterbot"
channel="#matterbridge"
UseSlackCircumfix=false
RemoteNickFormat="<{NICK}> "

[mattermost]
server="yourmattermostserver.domain"
team="yourteam"
#login/pass of your bot. Use a dedicated user for this and not your own!
login="yourlogin"
password="yourpass"
showjoinpart=true
channel="thechannel"
#whether to prefix messages from IRC to mattermost with the sender's nick. Useful if username overrides for incoming webhooks isn't enabled on the mattermost server
PrefixMessagesWithNick=false
#how to format the list of IRC nicks when displayed in mattermost. Possible options are "table" and "plain"
NickFormatter=plain
#how many nicks to list per row for formatters that support this
NicksPerRow=4
#Freenode nickserv
NickServNick="nickserv"
#Password for nickserv
NickServPassword="secret"
RemoteNickFormat="`irc` <{NICK}>"

[general]
#request your API key on https://github.com/giphy/GiphyAPI. This is a public beta key
GiphyApiKey="dc6zaTOxFJmzC"

#multiple channel config
[channel "our testing channel"]
irc="#bottesting"
mattermost="testing"

[channel "random channel"]
irc="#random"
mattermost="random"
```
