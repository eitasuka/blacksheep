/*
Copyright 2019 tira

This program is free software: you can redistribute it and/or modify it under the terms of the GNU
General Public License as published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without
even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not,
see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/chzyer/readline"
)

var (
	completer = readline.NewPrefixCompleter(
		readline.PcItem("clear"),
		readline.PcItem("help"),
		readline.PcItem("say"),
		readline.PcItem("list",
			readline.PcItem("channels"),
		),
		readline.PcItem("set",
			readline.PcItem("channel"),
		),
		readline.PcItem("channel"),
	)
	channelOptions map[string]string
)

// ControlAccount gives the user a console to interact with.
func ControlAccount(Discord *discordgo.Session, serverID string) {
	/*
	 * Write all the existing channels to the map.
	 */
	channelOptions = make(map[string]string)
	channels, err := Discord.GuildChannels(serverID)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			Fatal("No server provided to join.")
		}
		Fatal(err.Error())
	}
	for _, channel := range channels {
		channelOptions[channel.Name] = channel.ID
	}
	/*
	 * Using readline here makes the experience much more comfortable.
	 * It allows us to use many common cli features, like pressing up to run the
	 * previous command and such.
	 */
	reader, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
	})
	if err != nil {
		Fatal(err.Error())
	}
	defer reader.Close()
	var channel string
	/*
	 * Keep reading lines until the user exits.
	 */
	for {
		line, err := reader.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		switch {
		/*
		 * See what command line contains
		 */
		case strings.HasPrefix(line, "set "):
			switch {
			case strings.HasPrefix(line[4:], "channel"):
				/*
				 * Here we use the string the user entered as the key to a map, which
				 * contains the channel ID as a value.
				 */
				newChannel, ok := channelOptions[line[12:]]
				if !ok {
					Warning("Invalid channel name")
				} else {
					channel = newChannel
				}
			default:
				fmt.Println("Bad argument")
			}
		case strings.HasPrefix(line, "say "):
			_, err = Discord.ChannelMessageSend(channel, line[4:])
			if err != nil {
				if strings.Contains(err.Error(), "404") {
					Warning("You aren't in a channel.")
				} else if strings.Contains(err.Error(), "50001") {
					Warning("You don't have permission to talk here!")
				} else {
					Warning(err.Error())
				}
			}
		case strings.HasPrefix(line, "list "):
			switch {
			case line[5:] == "channels":
				/*
				 * Channels() also rebuilds the channel map, which adds the neat feature
				 * of updating when new channels are created.
				 */
				Channels(Discord, serverID)
			default:
				fmt.Println("Bad argument")
			}
		case line == "channel":
			fmt.Println(channel)
		case line == "help":
			Usage(reader.Stderr())
		case line == "clear":
			readline.ClearScreen(reader.Stdout())
		case line == "exit":
			return
		default:
			fmt.Println("?")
		}
	}
}

// Usage prints all the available commands.
func Usage(writer io.Writer) {
	io.WriteString(writer, "Commands:\n")
	io.WriteString(writer, completer.Tree("    "))
}

// Channels lists all the of channels in a given server.
func Channels(Discord *discordgo.Session, serverID string) {
	/*
	 * Collects every text channel in the set server and adds them as valid
	 * options for the set channel command.
	 */
	channels, err := Discord.GuildChannels(serverID)
	if err != nil {
		Warning(err.Error())
	}
	/* Initialising a map clears it */
	channelOptions = make(map[string]string)
	for _, channel := range channels {
		if channel.Type == 0 { /* type 0 is a text channel */
			fmt.Println(channel.Name)
			channelOptions[channel.Name] = channel.ID
		}
	}
}
