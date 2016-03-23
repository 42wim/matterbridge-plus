# matterbridge-plus

Simple bridge between mattermost and IRC. (Uses the mattermost API instead of webhooks)

Relays public channel messages between mattermost and IRC.
Supports multiple mattermost and irc channels.

Requires mattermost 2.1.0 and a dedicated user(bot) on your mattermost instance.
Matterbridge-plus also works with private groups.

There is also a version with webhooks that doesn't need a dedicated user. See [matterbridge] (https://github.com/42wim/matterbridge/)

## binaries
Binaries will be found [here] (https://github.com/42wim/matterbridge-plus/releases/tag/v0.1)

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
* send messages from mattermost to irc and vice versa, messages in mattermost will appear with irc-nick

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

[mattermost]
server="yourmattermostserver.domain"
team="yourteam"
#login/pass of your bot. Use a dedicated user for this and not your own!
login="yourlogin"
password="yourpass"
showjoinpart=true
channel="thechannel"

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
