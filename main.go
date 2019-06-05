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
/* https://lkml.org/lkml/2009/12/17/229 */

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	/*
	 * CLI arg parsing
	 */
	/*
	 * Creating a new kingpin parsing instance
	 */
	blacksheep = kingpin.New("blacksheep", "The Discord tooling powerhouse")
	/*
	 * serverID isn't required for every subcommand, but it's better to declare it here and check it
	 * if we need it than to have scrapeServerID, controlServerID, etc.
	 */
	serverID = kingpin.Flag("server", "A server ID to connect or scrape.").String()
	/*
	 * scrape scrapes content from a specified server and optionally, a specific channel. If no
	 * channel is provided, it downloads from every channel the account token has access to.
	 * Optionally, the type of content can be specified, either images or text logs. The default is
	 * both.
	 */
	scrape    = kingpin.Command("scrape", "Scrape images, logs, or both.")
	channelID = scrape.Flag("channel", "The channel ID within the server that is to be scraped."+
		"Default: all").String()
	/*
	 * A collection of standalone or simple commands to gather information about servers and users.
	 */
	channels = kingpin.Command("channels", "A list of channels in the provided server.")
	guild    = kingpin.Command("guild", "List of information about a provided server.")
	/*
	 * control gives the user the ability to control any account via a CLI.
	 */
	control = kingpin.Command("control", "Control an account.")
	/*
	 * self starts a selfbot instance.
	 */
	self = kingpin.Command("self", "Start a selfbot.")
)

// UserConfig is used across multiple files, so its easier to define it here.
var UserConfig Config

// Config is a struct used to store values taken from config.toml, and is used to generate a new
// config in the case of one not existing.
type Config struct {
	Token             string
	SaveDirectory     string
	SelfBotPrefix     string
	SelfBotCopypastas []string `toml:"Copypastas"`
}

func main() {
	command := kingpin.Parse()
	/*
	 * ParseConfig doesn't return anything because it modifies the module-level UserConfig variable.
	 */
	ParseConfig()
	if len(UserConfig.SelfBotCopypastas) > 0 {
		fmt.Printf("Custom copypastas: %+q\n", UserConfig.SelfBotCopypastas)
	}
	/*
	 * CreateDiscordInstance is located in auto.go, and simply returns a
	 * *discordgo.Session instance used in most all other functions.
	 */
	Discord, err := CreateDiscordInstance(UserConfig.Token)
	username, err := Discord.User("@me")
	if err != nil {
		if strings.Contains(err.Error(), "401: Unauthorized") {
			Fatal("Failed to connect to Discord. Your token may be invalid.")
		}
		Fatal("Failed to connect to Discord: " + err.Error())
	}
	Success(fmt.Sprintf("Connected as %v", username))
	switch command {
	case scrape.FullCommand():
		Scrape(Discord, *serverID, *channelID, UserConfig.SaveDirectory)
	case channels.FullCommand():
		GetChannelList(Discord, *serverID)
	case guild.FullCommand():
		GetGuildDetails(Discord, *serverID)
	case control.FullCommand():
		ControlAccount(Discord, *serverID)
	case self.FullCommand():
		SelfBot(Discord)
	}
}

// ParseConfig parses the blacksheep.toml file if it eists. Otherwise, it generates a new one from
// a template (the Config struct) and exits with an error code.
func ParseConfig() {
	if _, err := os.Stat(configDirectory); os.IsNotExist(err) {
		Warning("No blacksheep.toml found, generating a new one at " + configDirectory)
		newConfig := new(bytes.Buffer)
		if err := toml.NewEncoder(newConfig).Encode(UserConfig); err != nil {
			Fatal("Failed to encode a new config: " + err.Error())
		}
		if _, err := os.Stat(configFolder); os.IsNotExist(err) {
			os.Mkdir(configFolder, 0700)
		}
		file, err := os.OpenFile(configDirectory, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			Fatal("Failed to create blacksheep.toml: " + err.Error())
		}
		defer file.Close()
		fmt.Fprintf(file, newConfig.String())
		fmt.Println("Wrote blacksheep.toml. Edit this, then run Blacksheep again.")
		os.Exit(1)
	} else if err != nil {
		/*
		 * There was some other error with finding blacksheep.toml, maybe we don't have the right
		 * permissions?
		 */
		Fatal("Error with finding blacksheep.toml: " + err.Error())
	}
	/*
	 * blacksheep.toml exists.
	 */
	_, err := toml.DecodeFile(configDirectory, &UserConfig)
	if err != nil {
		/*
		 * This will probably only occur if the TOML data is invalidated by the user.
		 */
		Fatal("Failed to decode blacksheep.toml: " + err.Error())
	}
	/*
	 * If a custom SaveDirectory was defined, and it doesn't end with a trailing /, append one.
	 */
	if UserConfig.SaveDirectory != "" &&
		UserConfig.SaveDirectory[len(UserConfig.SaveDirectory)-1] != '/' {
		UserConfig.SaveDirectory = UserConfig.SaveDirectory + "/"
	}
	UserConfig.SaveDirectory = UserConfig.SaveDirectory + *serverID + "/"
	/*
	 * If no custom selfbot prefix is defined, we default it to ::
	 */
	if UserConfig.SelfBotPrefix == "" {
		UserConfig.SelfBotPrefix = "::"
	}
}
