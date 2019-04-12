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
/*
 * All of the commands used by selfbot.go
 */
package main

import (
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	charmap = map[string]string{
		"!": ":exclamation:",
		"?": ":question:",
		"+": ":heavy_plus_sign:",
		"-": ":heavy_minus_sign:",
		"£": ":pound:",
		"¥": ":yen:",
		"€": ":euro:",
	}
	IsLetter   = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
	copypastas = []string{
		`
    ⠀⠀⠀⠀⠀⠀⣤⣤
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

/*
 * the `command` command for selfbot.go
 */
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

func NewCustomCommand(command string) {

}

/*
 * the `huge` command for selfbot.go
 */

func Huge(str string) string {
	var hugeString strings.Builder
	for _, char := range str {
		if IsLetter(string(char)) {
			hugeString.WriteString(":regional_indicator_" + string(char) + ": ")
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

/*
 * The `copypasta` command for selfbot.go
 */

func Copypasta(custom []string) string {
	copypastas = append(copypastas, custom...)
	rand.Seed(time.Now().Unix())
	return copypastas[rand.Intn(len(copypastas))]
}

/*
 * the `help` command for selfbot.go
 */

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
			Value:  "Select a random copypasta",
			Inline: true,
		},
	}
}
