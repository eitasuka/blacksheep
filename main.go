/*
Copyright 2019 tira

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
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
	 * All the the tools use a server ID, so we might as well make it a base
	 * requirement. The only side-effect here is that kingpin will give an error
	 * about this not being provided when running ./blacksheep,
	 * ./blacksheep --help is required to get the actual help text.
	 */
	serverID = kingpin.Flag("server", "A server ID to connect or scrape.").
			String()
	/*
	 * scrape scrapes content from a specified server and optionally, a
	 * specific channel. If no channel is provided, it downloads from every
	 * channel the account token has access to. Optionally, the type of content
	 * can be specified, either images or text logs. The default is both.
	 */
	scrape    = kingpin.Command("scrape", "Scrape images, logs, or both.")
	channelID = scrape.Flag("channel", "The channel ID within the server that"+
		" is to be scraped. Default: all").String()
	media = scrape.Flag("media", "Save media only.").Bool()
	logs  = scrape.Flag("logs", "Save text chat logs only.").Bool()
	/*
	 * A collection of standalone or simple commands to gather information about
	 * servers and users.
	 */
	channels = kingpin.Command("channels", "A list of channels in the provided"+
		" server.")
	guild = kingpin.Command("guild", "List of information about a provided"+
		" server.")
	/*
	 * control gives the user the ability to control any account via a CLI.
	 */
	control = kingpin.Command("control", "Control an account.")
	/*
	 * self starts a selfbot instance.
	 */
	self = kingpin.Command("self", "Start a selfbot.")
)

type Config struct {
	/*
	 * Config is a struct used to store values taken from config.toml, and is
	 * used to generate a new config in the case of one not existing.
	 */
	Token             string
	SaveDirectory     string
	SelfBotPrefix     string
	SelfBotCopypastas []string
}

var UserConfig = ParseConfig()

func main() {
	command := kingpin.Parse()
	/*
	 * Here we create new *discordgo.Session instance. Discord will be passed
	 * around to most functions. An error here may not denote an invalid token,
	 * but network errors.
	 */
	Discord, err := CreateDiscordInstance(UserConfig.Token)
	if err != nil {
		Fatal(err.Error())
	}
	/*
	 * Here we validate the token given by making sure we can access @me.
	 */
	username, err := Discord.User("@me")
	if err != nil {
		if strings.Contains(err.Error(), "401: Unauthorized") {
			Fatal("HTTP 401 Unauthorized. Your token is probably invalid.")
		}
		Fatal(err.Error())
	}
	Success(fmt.Sprintf("Connected as %s", username))
	/*
	 * The rest of the code here parses the given arguments and runs the
	 * appropriate functions.
	 */
	switch command {
	/*
	 * All of the following functions, until otherwise stated, are located in
	 * auto.go
	 */
	case scrape.FullCommand():
		/*
		 * Try to create a directory to put all the log files in.
		 */
		if _, err := os.Stat(UserConfig.SaveDirectory +
			*serverID); os.IsNotExist(err) || err == nil {
			os.Mkdir(UserConfig.SaveDirectory+*serverID, 0700)
		} else {
			Fatal(err.Error())
		}
		if *media && !*logs {
			GetMedia(Discord, *channelID, *serverID, UserConfig.SaveDirectory)
		} else if *logs && !*media {
			GetLogs(Discord, *channelID, *serverID, UserConfig.SaveDirectory)
		} else {
			GetMediaAndLogs(Discord, *channelID, *serverID, UserConfig.SaveDirectory)
		}
	case channels.FullCommand():
		GetChannelList(Discord, *serverID)
	case guild.FullCommand():
		GetGuildDetails(Discord, *serverID)
	case control.FullCommand():
		ControlAccount(Discord, *serverID)
	case self.FullCommand():
		/*
		 * All the work here is done by code in the selfbot/ directory.
		 */
		Selfbot(Discord)
	}
}

func ParseConfig() Config {
	/*
	 * ParseConfig parses the blacksheep.toml file IF it exists. Otherwise, it
	 * generates a new one from a template and exits with an error code.
	 * ParseConfig relies on configDirectory, which is defined in separate source
	 * files depending on the target platform. On windows, parseconfig_windows.go
	 * is compiled, and the config directory is set to the current working
	 * directory, and parseconfig_unix.go is ignored. on !windows, it's
	 * reversed, and the config directory is set to
	 * ~/.config/blacksheep/blacksheep.toml
	 */
	var UserConfig Config
	_, err := toml.DecodeFile(configDirectory, &UserConfig)
	if err != nil {
		/*
		 * Something went wrong with parsing blacksheep.toml. Presumably, it doesn't
		 * Exist yet.
		 */
		if _, err := os.Stat(configDirectory); os.IsNotExist(err) {
			/* Warn the user that no (valid?) configuration file was found */
			Warning("No blacksheep.toml found, generating a new one at " +
				configDirectory)
			/*
			 * Encode the Config struct to TOML data
			 */
			document := new(bytes.Buffer)
			if err := toml.NewEncoder(document).Encode(UserConfig); err != nil {
				Fatal(err.Error())
			}
			/*
			 * If the blacksheep directory in configDirectory doesn't exist, make it.
			 */
			if _, err := os.Stat(configFolder); os.IsNotExist(err) {
				os.Mkdir(configFolder, 0700)
			}
			/*
			 * Make a new file at configDirectory and write the encoded data to it.
			 */
			file, err := os.OpenFile(configDirectory, os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				Fatal(err.Error())
			}
			/*
			 * Write the data using fmt.Fprintf to the file's io.Writer,
			 * inform the user, then exit.
			 */
			defer file.Close()
			fmt.Fprintf(file, document.String())
			fmt.Println("Wrote blacksheep.toml. Edit this,then run BlackSheep" +
				" again.")
			os.Exit(1)
		}
	}
	/*
	 * Have I written UserConfig.SaveDirectory enough for you?
	 *
	 * Here, SaveDirectory has a trailing / appended if there isn't one already.
	 */
	if UserConfig.SaveDirectory != "" &&
		UserConfig.SaveDirectory[len(UserConfig.SaveDirectory)-1] != '/' {
		UserConfig.SaveDirectory = UserConfig.SaveDirectory + "/"
	}
	/*
	 * If the user hasn't defined a custom selfbot prefix, we default to ::
	 */
	if UserConfig.SelfBotPrefix == "" {
		UserConfig.SelfBotPrefix = "::"
	}
	return UserConfig
}
