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
	"os"
	"strings"
	"text/tabwriter"

	"github.com/bwmarrin/discordgo"
)

// CreateDiscordInstance is used by main to create a new Discord session.
func CreateDiscordInstance(Token string) (*discordgo.Session, error) {
	return discordgo.New(Token)
}

// GetGuildDetails returns useful information about a guild, including the name, id, owners id,
// region, and available roles.
func GetGuildDetails(Discord *discordgo.Session, serverID string) {
	guild, err := Discord.Guild(serverID)
	if err != nil {
		Fatal(err.Error())
	}
	fmt.Printf("ID | %v\nName | %v\nRegion | %v\nOwner ID | %v\n",
		guild.ID, guild.Name, guild.Region, guild.OwnerID)
	/*
	 * Spliting this across a few print calls to make it a *touch* prettier.
	 */
	fmt.Printf("AFK Timeout | %v\nVerfication Level | %v\nEmbedding | %t\n",
		guild.AfkTimeout, guild.VerificationLevel, guild.EmbedEnabled)
	fmt.Printf("Content Filter Level | %v\n", guild.ExplicitContentFilter)
	/*
	 * Go through every role and print the information that matters about them, cleanly.
	 */
	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug|tabwriter.AlignRight)
	fmt.Fprintln(writer, "ROLE\tMANAGED\tHOISTED\tCOLOR\tPOSITION")
	for _, role := range guild.Roles {
		fmt.Fprintln(writer, strings.Replace(fmt.Sprintf("%v\t %t\t %v\t %v\t %v",
			role.Name, role.Managed, role.Hoist, role.Color, role.Position), "\"", "", -1))
	}
	writer.Flush()
}

// GetChannelList gets a list of (accessable given the users permissions) channels in a specified
// server.
func GetChannelList(Discord *discordgo.Session, serverID string) {
	guild, err := Discord.Guild(serverID)
	channels, err := Discord.GuildChannels(serverID)
	if err != nil {
		Fatal(err.Error())
	}
	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug|tabwriter.AlignRight)
	fmt.Fprintf(writer, "Channel list for %v (%v)\n", guild.Name, guild.ID)
	fmt.Fprintln(writer, "ID\tNAME\tTYPE\tNSFW\t")
	for _, channel := range channels {
		fmt.Fprintln(writer, strings.Replace(fmt.Sprintf("%v\t %v\t %v\t %t",
			channel.ID, channel.Name, ChannelType(channel.Type), channel.NSFW), "\"", "", -1))
	}
	writer.Flush()
}

// ChannelType returns the type that a channel is, as DiscordGo only provides a numerical id
// with an associated type.
func ChannelType(channelType discordgo.ChannelType) string {
	switch channelType {
	case 0:
		return "Guild Text"
	case 1:
		return "DM"
	case 2:
		return "Guild Voice"
	case 3:
		return "Group DM"
	case 4:
		return "Guild Category"
	}
	return "Unknown" /* Hva??? */
}
