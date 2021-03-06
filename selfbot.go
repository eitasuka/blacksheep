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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	customCommands = make(map[string]string)
	charmap        = map[string]string{
		"!": ":exclamation:",
		"?": ":question:",
		"+": ":heavy_plus_sign:",
		"-": ":heavy_minus_sign:",
		"£": ":pound:",
		"¥": ":yen:",
		"€": ":euro:",
	}
	// IsLetter affirms that a provided string is indeed... a letter.
	IsLetter = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
	owoify   = strings.NewReplacer(
		"you", "u",
		"love", "luv",
		"r", "w",
		"l", "w",
		"n", "ny",
		" to", "a",
		"and", "n",
		".", " uwu ",
		":)", "nyaa!~")
)

const (
	colorError   = 0x8b1117 /* Dark red */
	colorSuccess = 0x006606 /* Dark green */
	colorNotice  = 0x00549d /* Dark blue */
)

// SelfBot parses the available custom commands and creates a handler that calls
// OnMessageCreate.
func SelfBot(Session *discordgo.Session) {
	/*
	 * What it assigns the handler to is based on what arguments it takes, this is assigned
	 * to fire whenever a new message is sent.
	 */
	if !*noNew {
		Session.AddHandler(LogMessageNew)
	}
	Session.AddHandler(LogMessageUpdate)
	Session.AddHandler(LogMessageDelete)
	/*
	 * ParseCustomCommands() uses json.Unmashal to pass the data of commands.json
	 * directly to the customCommands map.
	 */
	ParseCustomCommands()
	Session.AddHandler(OnMessageCreate)
	err := Session.Open()
	if err != nil {
		Fatal("Failed to start selfbot, " + err.Error())
	}
	fmt.Println("Using prefix", UserConfig.SelfBotPrefix)
	fmt.Println("Listening for messages")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-signalChan
}

// OnMessageCreate parses every message it finds, and if applicable, parses the user's command.
func OnMessageCreate(Session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID != Session.State.User.ID ||
		!strings.HasPrefix(message.Content, UserConfig.SelfBotPrefix) {
		/*
		 * Quickly ignore all messages not send by the bot owner, and messages that
		 * don't start with the bot's identifier.
		 */
		return
	}
	command := strings.Split(
		message.Content[len(UserConfig.SelfBotPrefix):len(message.Content)], " ")[0]
	var messageContent string
	if len(strings.Split(message.Content, " ")) > 1 {
		messageContent = message.Content[len(UserConfig.SelfBotPrefix)+
			len(command)+1 : len(message.Content)]
	}
	newMessage := discordgo.NewMessageEdit(message.ChannelID, message.ID)
	/*
	 * We only want to display the embed here, so we empty out the message
	 * "content".
	 */
	newMessage.SetContent("")
	switch command {
	case "about":
		newMessage.SetEmbed(&discordgo.MessageEmbed{
			URL:         "https://github.com/t1ra/blacksheep",
			Title:       "Blacksheep",
			Description: "The Discord tooling powerhouse.",
			Color:       colorNotice,
		})
	case "help":
		newMessage.SetEmbed(&discordgo.MessageEmbed{
			Title:       "Blacksheep",
			Description: "Help",
			Color:       colorNotice,
			Fields:      append(HelpFields(), CustomCommands(customCommands)...),
		})
	case "details":
		newMessage.SetEmbed(Details(Session, message))
	case "avatar":
		newMessage.SetEmbed(Avatar(Session, message))
	case "huge":
		newMessage.SetContent(Huge(messageContent))
	case "copypasta":
		newMessage.SetContent(Copypasta(UserConfig.SelfBotCopypastas))
	case "command":
		switch strings.Split(messageContent, " ")[0] {
		case "new":
			newMessage.SetEmbed(NewCustomCommand(messageContent))
		case "delete":
			newMessage.SetEmbed(DeleteCustomCommand(messageContent))
		default:
			if !UserConfig.Lowkey {
				newMessage.SetEmbed(&discordgo.MessageEmbed{
					Title:       "Hva?",
					Description: "That doesn't look like a valid sub-command.",
					Color:       colorError,
				})
			}
		}
	case "owoify":
		newMessage.SetContent(Owoify(messageContent))
	case "epoch":
		newMessage.SetContent(Epoch())
	case "spam":
		newMessage.SetContent(Spam(messageContent, true))
	case "spamns":
		newMessage.SetContent(Spam(messageContent, false))
	case "everyone":
		newMessage.SetContent(TagEveryone(Session, message))
	default:
		if content, ok := customCommands[command]; ok {
			newMessage.SetContent(content)
		} else {
			if !UserConfig.Lowkey {
				newMessage.SetEmbed(&discordgo.MessageEmbed{
					Title:       "Hva?",
					Description: "That doesn't look like a valid command.",
					Color:       colorError,
				})
			}
		}
	}
	/*
	 * Edit the message containing the command, replacing its contents with
	 * whatever was generated by the switch.
	 */
	Session.ChannelMessageEditComplex(newMessage)
}

/*
 * The functions associated with the parsed commmand.
 */

// CustomCommands returns a list of user-defined commands, which is appended to the default Help()
// text.
func CustomCommands(commands map[string]string) []*discordgo.MessageEmbedField {
	var embed []*discordgo.MessageEmbedField
	for key, value := range commands {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   key,
			Value:  value,
			Inline: true,
		})
	}
	return embed
}

// ParseCustomCommands adds every custom command found in commands.json to a variable.
func ParseCustomCommands() {
	commandsFile := configFolder + "commands.json"
	if _, err := os.Stat(commandsFile); os.IsNotExist(err) {
		file, err := os.Create(commandsFile)
		if err != nil {
			Fatal("Failed to create commands.json: " + err.Error())
		}
		file.Close()
		return
	}
	data, err := ioutil.ReadFile(commandsFile)
	if string(data) == "" {
		return
	}
	if err != nil {
		Fatal("Failed to open commands.json: " + err.Error())
	}
	err = json.Unmarshal(data, &customCommands)
	if err != nil {
		Fatal("Failed to parse commands.json: " + err.Error())
	}
	if len(customCommands) != 0 {
		fmt.Println("Loaded custom commands:")
		for key := range customCommands {
			fmt.Println(key)
		}
	}
}

// NewCustomCommand creates a new custom command by adding it to the custom commands variable,
// and to commands.json.
func NewCustomCommand(command string) *discordgo.MessageEmbed {
	if len(command) < 5 {
		return &discordgo.MessageEmbed{
			Title:       "Failed to create new custom command",
			Description: "You haven't provided a command name.",
			Color:       colorError,
		}
	}
	command = command[4:len(command)]
	if len(strings.Split(command, " ")) <= 1 {
		return &discordgo.MessageEmbed{
			Title:       "Failed to create new custom command",
			Description: "You haven't provided a command body.",
			Color:       colorError,
		}
	}
	commandSplit := strings.Split(command, " ")
	commandName := commandSplit[0]
	commandBody := strings.Join(commandSplit[1:len(commandSplit)], " ")
	customCommands[commandName] = commandBody
	jsonData, err := json.Marshal(customCommands)
	if err != nil {
		Fatal("Failed to marshal custom commands: " + err.Error())
	}
	ioutil.WriteFile(configFolder+"commands.json", []byte(jsonData), 0600)
	return &discordgo.MessageEmbed{
		Title:       "Created new custom command",
		Description: fmt.Sprintf("Created new custom command %v", commandName),
		Color:       colorSuccess,
	}
}

// DeleteCustomCommand deletes a custom command from both the custom commands variable, and
// commands.json.
func DeleteCustomCommand(command string) *discordgo.MessageEmbed {
	if len(strings.Split(command, " ")) == 1 {
		return &discordgo.MessageEmbed{
			Title:       "Failed to delete custom command",
			Description: "No command was provided to delete.",
			Color:       colorError,
		}
	}
	command = strings.Split(command, " ")[1]
	if _, exists := customCommands[command]; !exists {
		return &discordgo.MessageEmbed{
			Title:       "Failed to delete custom command",
			Description: "That command doesn't exist.",
			Color:       colorError,
		}
	}
	delete(customCommands, command)
	jsonData, err := json.Marshal(customCommands)
	if err != nil {
		Fatal("Failed to marshal custom commands: " + err.Error())
	}
	ioutil.WriteFile(configFolder+"commands.json", []byte(jsonData), 0600)
	return &discordgo.MessageEmbed{
		Title:       "Deleted custom command",
		Description: fmt.Sprintf("Deleted %v", command),
		Color:       colorSuccess,
	}
}

// Huge turns every letter given to it to a regional indicator, and a few other special characters
// that discord supports.
func Huge(str string) string {
	var hugeString strings.Builder
	for _, char := range str {
		if IsLetter(string(char)) {
			hugeString.WriteString(":regional_indicator_" + strings.ToLower(string(char)) + ": ")
		} else if val, ok := charmap[string(char)]; ok {
			hugeString.WriteString(val + " ")
		} else if string(char) == " " {
			hugeString.WriteString("  ")
		} else if string(char) == "\n" {
			hugeString.WriteString("\n")
		}
	}
	if hugeString.String() == "" {
		return str
	}
	return hugeString.String()
}

// Copypasta returns a super randomTM copypasta from the copypastas variable.
func Copypasta(custom []string) string {
	copypastas = append(copypastas, custom...)
	rand.Seed(time.Now().Unix())
	return copypastas[rand.Intn(len(copypastas)-1)+1]
}

// HelpFields returns every help item for every command in Blacksheep.
func HelpFields() []*discordgo.MessageEmbedField {
	return []*discordgo.MessageEmbedField{
		&discordgo.MessageEmbedField{
			Name:   "help",
			Value:  "Show this embed.",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "about",
			Value:  "Show an about Blacksheep embed.",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "huge",
			Value:  "Convert a string into emojis, e.g. regional indicators.",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "copypasta",
			Value:  "Select a random copypasta.",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name: "command",
			Value: "Create (command new <name> <body>) or delete (command" +
				" delete <name>) custom commands.",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "owoify",
			Value:  "OwO-ify a string! nyaa~",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "details",
			Value:  "Get details about a certain user.",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "avatar",
			Value:  "Get a users avatar.",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "epoch",
			Value:  "Get the current epoch.",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "spam",
			Value:  "Spam a provided string. Or e.",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "spamns",
			Value:  "Same as spam, but without space delimiters",
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "everyone",
			Value:  "Manually tag every user in the server (if you can't use @everyone)",
			Inline: true,
		},
	}
}

// Owoify returns a string after passing it through a simple find-and-replace function
// defined above in var().
func Owoify(message string) string {
	return owoify.Replace(message)
}

// Details returns details of a tagged user. If no user is tagged, it returns the bot owner's
// details, like their profile picture, name, and such.
func Details(Session *discordgo.Session, message *discordgo.MessageCreate) *discordgo.MessageEmbed {
	User := message.Author
	if len(message.Mentions) > 0 {
		User = message.Mentions[0]
	}
	return &discordgo.MessageEmbed{
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: User.AvatarURL("512"),
		},
		Title:       User.Username + "#" + User.Discriminator,
		Description: "ID: " + User.ID,
		Color:       colorSuccess,
	}
}

// Avatar returns the avatar of a tagged user. If no user is tagged, it returns the bot owner's
// avatar.
func Avatar(Session *discordgo.Session, message *discordgo.MessageCreate) *discordgo.MessageEmbed {
	User := message.Author
	if len(message.Mentions) > 0 {
		User = message.Mentions[0]
	}
	return &discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: User.AvatarURL("512"),
		},
		Color: colorSuccess,
	}
}

// Epoch returns the current UNIX epoch.
func Epoch() string {
	return string(strconv.FormatInt(time.Now().Unix(), 10))
}

// Spam spams a provided letter (or e) the maximum amount of times.
func Spam(str string, space bool) string {
	if len(str) == 0 {
		str = "e"
	}
	var nyString strings.Builder
	if space {
		for nyString.Len() < 2000 {
			nyString.WriteString(str + " ")
		}
	} else {
		for nyString.Len() < 2000 {
			nyString.WriteString(str)
		}
	}
	/* Just in case we're spamming an :emote:, it's better to cut it
	 * at the last emote rather than splitting it.
	 */
	if string(str[0]) == ":" {
		return nyString.String()[0:strings.LastIndex(nyString.String()[:2000], ":")]
	}
	return nyString.String()[:2000]
}

// IDToTimestamp converts a Discord message id "snowflake" to a timestamp with some
// simple maths. Stolen from https://github.com/vegeta897/snow-stamp.
func IDToTimestamp(id string) time.Time {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		// This like, won't happen.
		os.Exit(-8008)
	}
	return time.Unix((int64(idInt)/4194304)+1420070400000, 0)
}

// TagEveryone manually @'s every person in a server.
func TagEveryone(Session *discordgo.Session, m *discordgo.MessageCreate) string {
	guild := m.GuildID
	members, err := Session.GuildMembers(guild, "", 1000)

	if err != nil {
		Fatal("Failed to get guild members. Bad connection? " + err.Error())
	}

	var builder strings.Builder
	for _, member := range members {
		if member.User.ID == m.Message.Author.ID {
			// Don't waste space tagging ourselves.
			continue
		}
		fmt.Fprintf(&builder, member.Mention())
		fmt.Fprintf(&builder, " ")
	}
	return builder.String()
}

// LogMessageNew logs message creation events.
func LogMessageNew(Session *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("At %v, in %v/%v, %v said:\n%v\n", IDToTimestamp(m.Message.ID), m.GuildID,
		m.ChannelID, m.Message.Author, m.Message.Content)
}

// LogMessageUpdate logs message edit events.
func LogMessageUpdate(Session *discordgo.Session, m *discordgo.MessageUpdate) {
	fmt.Printf("At %v, in %v/%v, %v updated a message to say: %v\n",
		IDToTimestamp(m.Message.ID), m.Message.GuildID, m.Message.ChannelID, m.Message.Author,
		m.Message.Content)
}

// LogMessageDelete logs message deletion events.
func LogMessageDelete(Session *discordgo.Session, m *discordgo.MessageDelete) {
	fmt.Printf("At %v, in %v/%v, %v deleted a message saying: %v\n",
		IDToTimestamp(m.Message.ID), m.Message.GuildID, m.Message.ChannelID, m.Message.Author,
		m.Message.Content)
}

/* I'll leave this at the bottom because its unsightly. */
var (
	copypastas = []string{
		`
		⠀⠀ ⣤⣤
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿
⠀⠀⣶⠀⠀⣀⣤⣶⣤⣉⣿⣿⣤⣀
⠤⣤⣿⣤⣿⠿⠿⣿⣿⣿⣿⣿⣿⣿⣿⣀
⠀⠛⠿⠀⠀⠀⠀⠉⣿⣿⣿⣿⣿⠉⠛⠿⣿⣤
⠀⠀⠀⠀⠀⠀⠀⠀⠿⣿⣿⣿⠛⠀⠀⠀⣶⠿
⠀⠀⠀⠀⠀⠀⠀⠀⣀⣿⣿⣿⣿⣤⠀⣿⠿
⠀⠀⠀⠀⠀⠀⠀⣶⣿⣿⣿⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠿⣿⣿⣿⣿⣿⠿⠉⠉
⠀⠀⠀⠀⠀⠀⠀⠉⣿⣿⣿⣿⠿
⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⠉
⠀⠀⠀⠀⠀⠀⠀⠀⣛⣿⣭⣶⣀
⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⠉⠛⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⠀⠀⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣉⠀⣶⠿
⠀⠀⠀⠀⠀⠀⠀⠀⣶⣿⠿
⠀⠀⠀⠀⠀⠀⠀⠛⠿⠛
	`, `
▒▒░░░░░░░░░░▄▐░░░░
▒░░░░░░▄▄▄░░▄██▄░░░
░░░░░░▐▀█▀▌░░░░▀█▄░
░░░░░░▐█▄█▌░░░░░░▀█▄
░░░░░░░▀▄▀░░░▄▄▄▄▄▀▀
░░░░░▄▄▄██▀▀▀▀░░░░░
░░░░█▀▄▄▄█░▀▀░░░░░░
░░░░▌░▄▄▄▐▌▀▀▀░░░░░
░▄░▐░░░▄▄░█░▀▀░░░░░
░▀█▌░░░▄░▀█▀░▀░░░░░
░░░░░░░░▄▄▐▌▄▄░░░░░
░░░░░░░░▀███▀█░▄░░░
░░░░░░░▐▌▀▄▀▄▀▐▄░░░
░░░░░░░▐▀░░░░░░▐▌░░
░░░░░░░█░░░░░░░░█░░
	`, `
If you're being mugged, just say no. Your robbers cannot legally take any of
your possessions.
	`, `
Do you seriously spend over 40 hours (not minutes) a week playing video games?
Can't tell if you're joking. If not then... wow. The only game I could play for
40 hours is nfl 18 and that's because I enjoy kicking other people's asses (I
have a natural instinct for strategy games and high-level thinking). But I
can't because I'm an adult unlike y'all I presume. Damn high school was fun :/
minus the detentions for picking on the nerds, haha.
	`,
	}
)
