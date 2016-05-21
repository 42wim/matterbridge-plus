# matterbridge-plus

Simple bridge between mattermost and IRC. (Uses the mattermost API instead of webhooks)

* Relays public channel messages between mattermost and IRC.
* Supports multiple mattermost and irc channels.
* Matterbridge-plus also works with private groups on your mattermost.

## Requirements:
* [Mattermost] (https://github.com/mattermost/platform/) 3.0.0 (stable, not a dev build)
* For older versions of mattermost 2.x use matterbridge-plus tag v0.2
* A dedicated user(bot) on your mattermost instance.

Master branch of matterbridge-plus should always work against latest STABLE mattermost release.
If you want to run matterbridge-plus with mattermost DEV builds, use the develop branch of matterbridge-plus

There is also a version with webhooks that doesn't need a dedicated user. See [matterbridge] (https://github.com/42wim/matterbridge/)   

## binaries
Binaries (for mattermost 3.0) can be found [here] (https://github.com/42wim/matterbridge-plus/releases/tag/v0.3)
Older binaries can be found [here] (https://github.com/42wim/matterbridge-plus/releases/)

## building
Go 1.6 is required
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
Usage of ./matterbridge-plus:
  -conf string
        config file (default "matterbridge.conf")
  -debug
        enable debug
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
#channel= is deprecated please use the channel config below
UseSlackCircumfix=false
#Freenode nickserv
NickServNick="nickserv"
#Password for nickserv
NickServPassword="secret"
RemoteNickFormat="<{NICK}> "
#Ignore the messages from these nicks. They will not be sent to mattermost
IgnoreNicks="ircspammer1 ircspammer2"

[mattermost]
server="yourmattermostserver.domain"
team="yourteam"
#login/pass of your bot. Use a dedicated user for this and not your own!
login="yourlogin"
password="yourpass"
showjoinpart=true
#channel= is deprecated please use the channel config below
#whether to prefix messages from IRC to mattermost with the sender's nick. Useful if username overrides for incoming webhooks isn't enabled on the mattermost server
PrefixMessagesWithNick=false
#how to format the list of IRC nicks when displayed in mattermost. Possible options are "table" and "plain"
NickFormatter=plain
#how many nicks to list per row for formatters that support this
NicksPerRow=4
RemoteNickFormat="`irc` <{NICK}>"
#Ignore the messages from these nicks. They will not be sent to irc
IgnoreNicks="mmbot spammer2"

[general]
#request your API key on https://github.com/giphy/GiphyAPI. This is a public beta key
GiphyApiKey="dc6zaTOxFJmzC"

#channel config
[channel "our testing channel"]
irc="#bottesting"
mattermost="testing"

[channel "random channel"]
irc="#random"
mattermost="random"
```
