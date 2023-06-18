// SPDX-License-Identifier: MIT

package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"log"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
)

type Flags struct {
	config_path *string
}

type Config struct {
	Channels    []string
	Debug       bool
	Meowchannel string
	Meowlow     int64
	Meowhigh    int64
	Meowreply   bool
	Nick        string
	Server      string
	Tls         bool
}

func add_callbacks(irc *ircevent.Connection, config *Config) {
	irc.AddConnectCallback(func(e ircmsg.Message) {
		if botMode := irc.ISupport()["BOT"]; botMode != "" {
			irc.Send("MODE", irc.CurrentNick(), "+"+botMode)
		}

		for _, channel := range config.Channels {
			irc.Join(channel)
		}
	})

	irc.AddCallback("PRIVMSG", func(e ircmsg.Message) {
		pleading_tomato_emoji(*irc, e, *config)
	})

	irc.AddCallback("PING", func(e ircmsg.Message) {
		if numgen(config.Meowlow, config.Meowhigh) == config.Meowlow {
			irc.Privmsg(config.Meowchannel, "meow")
		}
	})
}

func numgen(min int64, max int64) int64 {
	if max-min <= 0 {
		log.Print("max-min <= 0, returning min")
		return min
	}

	bigint, err := rand.Int(rand.Reader, big.NewInt(max-min))
	if err != nil {
		log.Fatal(err)
	}

	return bigint.Int64() + min
}

func pleading_tomato_emoji(irc ircevent.Connection, e ircmsg.Message, config Config) {
	text := strings.ToLower(e.Params[1])

	if strings.Contains(text, "\001action pats "+irc.CurrentNick()) {
		irc.Privmsg(e.Params[0], prr())
	}

	if strings.Contains(text, "mikobot cute") {
		irc.Privmsg(e.Params[0], not_cute())
	}

	if config.Meowreply == true {
		if strings.Contains(text, "meow") {
			irc.Privmsg(e.Params[0], "meow")
		}
	}
}

func not_cute() string {
	msg_len := numgen(8, 45)

	var msg string
	for i := 0; int64(i) < msg_len; i++ {
		msg += to_char(int(numgen(0, 84)))
	}

	return msg
}

func parse_flags() Flags {
	flags := Flags{
		config_path: flag.String("c", "/etc/mikobot/config.json", "path to config dir"),
	}

	flag.Parse()
	return flags
}

func prr() string {
	msg_len := numgen(6, 18)

	prrlet := make([]string, 3)
	prrlet[0] = "p"
	prrlet[1] = "r"
	prrlet[2] = "r"

	msg := "pr"
	for i := 0; int64(i) < msg_len; i++ {
		if strings.HasSuffix(msg, "p") {
			msg += prrlet[1]
		} else {
			msg += prrlet[numgen(0, 2)]
		}
	}

	return msg
}

func to_char(i int) string {
	return string(32 + i)
}

func main() {
	flags := parse_flags()
	config_json, err := os.ReadFile(*flags.config_path)
	var config Config
	err = json.Unmarshal(config_json, &config)

	if err != nil {
		log.Fatal(err)
	}

	irc := ircevent.Connection{
		Server: config.Server,
		Nick:   config.Nick,
		Debug:  config.Debug,
		UseTLS: config.Tls,
	}

	add_callbacks(&irc, &config)

	err = irc.Connect()
	if err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		irc.Quit()
	}()

	irc.Loop()
}
