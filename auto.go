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
 * auto.go contais functions that are automated -- scraping, detail grabbing,
 * and such.
 */
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/bwmarrin/discordgo"
	"mvdan.cc/xurls/v2"
)

var (
	/*
	 * Used by GetMediaFromChannel
	 */
	mediaTypes = []string{".png", ".jpg", ".gif", ".webm", ".mp4", ".mp3",
		".jpeg", ".jfif"}
)

func CreateDiscordInstance(Token string) (*discordgo.Session, error) {
	/*
	 * Since a lot of functions use a Discord instance, it's easier to create one
	 * in main(), then pass it to other functions as an argument. It returns a
	 * new *discordgo.Session instance.
	 */
	return discordgo.New(Token)
}

/*
 *
 */

func GetMediaAndLogs(Discord *discordgo.Session, channelID, serverID,
	saveDirectory string) {
	/*
	 * GetMediaAndLogs simply calls GetMedia and GetLogs sequentially.
	 */
	GetMedia(Discord, channelID, serverID, saveDirectory)
	GetLogs(Discord, channelID, serverID, saveDirectory)
}

func GetMedia(Discord *discordgo.Session, channelID, serverID,
	saveDirectory string) {
	/*
	 * GetImages doesn't do much compared to its internal functions. It mainly
	 * makes sure the channel to be scraped is both accessable and contains
	 * images, then creates a WaitGroup and calls one instance of
	 * GetMediaFromChannel for every channel being scraped.
	 */
	var wg sync.WaitGroup
	if channelID == "" {
		for _, channel := range AllChannels(Discord, serverID, 0) {
			wg.Add(1)
			go GetMediaFromChannel(Discord, channel, serverID, saveDirectory, &wg)
		}
	} else {
		wg.Add(1)
		go GetMediaFromChannel(Discord, channelID, serverID, saveDirectory, &wg)
	}
	wg.Wait()
}

func GetLogs(Discord *discordgo.Session, channelID, serverID,
	saveDirectory string) {
	/*
	 * GetLogs also doesn't do much compared to its internal functions. It mainly
	 * makes sure the channel to be scraped is both accessable and contains any
	 * messages, then creates a WaitGroup and calls one instance of
	 * GetLogsFromChannel for every channel being scraped.
	 */
	var wg sync.WaitGroup
	if channelID == "" {
		for _, channel := range AllChannels(Discord, serverID, 0) {
			wg.Add(1)
			go GetLogsFromChannel(Discord, channel, serverID, saveDirectory, &wg)
		}
	} else {
		wg.Add(1)
		go GetLogsFromChannel(Discord, channelID, serverID, saveDirectory, &wg)
	}
	wg.Wait()
}

/*
 * Internal functions for blacksheep scrape
 */

func DownloadMedia(url, channelID, serverID, saveDirectory string,
	wg *sync.WaitGroup) {
	/*
	 * DownloadMedia accepts a URL, creates a file in
	 * <serverid>/<channelid>/<medianame>, either relative to execution path if
	 * SaveDirectory isn't set in the config file, or
	 * <savedirectory>/<serverid>/<channelid>/<medianame> and saves the URL's
	 * data there.
	 */
	defer wg.Done()
	client := http.Client{
		Timeout: time.Duration(15 * time.Second),
	}
	response, err := client.Get(url)
	if err != nil {
		Warning("Timed out downloading " + url)
	}
	defer response.Body.Close()
	file, err := os.Create(saveDirectory + serverID + "/" + channelID + "/" +
		path.Base(url))
	if err != nil {
		Fatal("Failed to create file with error " + err.Error())
	}
	defer file.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		Warning("Failed to write image to file with error " + err.Error())
	}
}

func GetMediaFromChannel(Discord *discordgo.Session, channelID,
	serverID, saveDirectory string, wg *sync.WaitGroup) {
	/*
	 * GetImagesFromChannel is used internally as a goroutine. One instance of
	 * GetImagesFromChannel is called for every channel being scraped. It
	 * checks every message in the channel; if it contains an image, it
	 * saves it, otherwise, it throws it away.
	 */
	defer wg.Done()
	/*
	 * See if we can access the channel is empty, or some other issue arises.
	 */
	messages, err := Discord.ChannelMessages(channelID, 1, "", "", "")
	if len(messages) == 0 {
		Warning(channelID + " is empty.")
		return
	}
	if err != nil {
		Warning(fmt.Sprintf("Thread for channel %v failed with error %v!",
			channelID, err.Error()))
	}
	/*
	 * Try to create a folder at <serverid>/<channelid> to save the images to.
	 */
	if _, err := os.Stat(saveDirectory + serverID + "/" +
		channelID); os.IsNotExist(err) || err == nil {
		os.Mkdir(saveDirectory+serverID+"/"+channelID, 0700)
	} else {
		Fatal(err.Error())
	}
	/*
	 * Since ChannelMessages() can only get 100 messages at a time, we need to
	 * keep going in 100 message chunks until we reach the end.
	 */
	var mediaDownloadWaitGroup sync.WaitGroup
	for len(messages) != 0 {
		/*
		 * Discord.ChannelMessages will keep len(messages) > 0 for as long as its
		 * finding new messages in a said channel. We know we've hit the end of the
		 * channel when it's at 0.
		 */
		messages, err = Discord.ChannelMessages(channelID, 100,
			messages[len(messages)-1].ID, "", "")
		if err != nil {
			Warning(fmt.Sprintf("Thread for channel %v failed with error %v!",
				channelID, err.Error()))
		}
		/*
		 * This artisian code block first checks if the message contains any URLs,
		 * checks if those URLs have extensions that we care about, then goes
		 * through those and downloads them all in blocks of 100 (or however many
		 * messages were passed) as to avoid opening too many connections.
		 */
		for _, message := range messages {
			if message.Content != "" {
				urls := xurls.Strict().FindAllString(message.Content, -1)
				if len(urls) != 0 {
					for _, extension := range mediaTypes {
						for _, url := range urls {
							if strings.HasSuffix(strings.ToLower(url), extension) {
								mediaDownloadWaitGroup.Add(1)
								go DownloadMedia(url, channelID, serverID, saveDirectory,
									&mediaDownloadWaitGroup)
							}
						}
					}
				}
			} else if len(message.Attachments) != 0 {
				for _, attachment := range message.Attachments {
					mediaDownloadWaitGroup.Add(1)
					go DownloadMedia(attachment.URL, channelID, serverID, saveDirectory,
						&mediaDownloadWaitGroup)
				}
			} else if len(message.Embeds) != 0 {
				for _, embed := range message.Embeds {
					if embed.Image != nil {
						mediaDownloadWaitGroup.Add(1)
						go DownloadMedia(embed.Image.URL, channelID, serverID,
							saveDirectory, &mediaDownloadWaitGroup)
					}
				}
			}
			mediaDownloadWaitGroup.Wait()
		}
	}
	fmt.Printf("Finished scraping %v\n", channelID)
}

func GetLogsFromChannel(Discord *discordgo.Session, channelID, serverID,
	saveDirectory string, wg *sync.WaitGroup) {
	/*
	 * GetLogsFromChannel is used internally by GetLogs as a goroutine. One
	 * instance of GetLogsFromChannel is called for every channel being scraped.
	 * It checks every message in a channel in 100 message chunks, and adds them
	 * to a slice. Once there are no more messages, it writes the messages to
	 * a file with the name of the channel.
	 */
	defer wg.Done()
	/*
	 * See if we can access the channel is empty, or some other issue arises.
	 */
	messages, err := Discord.ChannelMessages(channelID, 1, "", "", "")
	if len(messages) == 0 {
		Warning(channelID + " is empty.")
		return
	}
	if err != nil {
		Warning(fmt.Sprintf("Thread for channel %v failed with error %v!",
			channelID, err.Error()))
	}
	/*
	 * Try to create a file at <serverid>/<channelid>.log to write the messages
	 * to.
	 */
	file, err := os.OpenFile(saveDirectory+serverID+"/"+channelID+".log",
		os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		Fatal(err.Error())
	}
	defer file.Close()
	/*
	 * Since ChannelMessages() can only get 100 messages at a time, we need to
	 * keep going in 100 message chunks until we reach the end.
	 */
	var allMessages []*discordgo.Message
	for len(messages) != 0 {
		messages, err = Discord.ChannelMessages(channelID, 100,
			messages[len(messages)-1].ID, "", "")
		if err != nil {
			Warning(fmt.Sprintf("Thread for channel %v failed with error %v!",
				channelID, err.Error()))
		}
		/*
		 * messages... appends all of the items *in* messages, rather than appending
		 * the slice itself.
		 */
		allMessages = append(allMessages, messages...)
	}
	/*
	 * Write all of the messages into a file.
	 */
	for _, message := range allMessages {
		file.WriteString(fmt.Sprintln(message.Author, "\n",
			message.Content+"\n\n"))
	}
	fmt.Printf("Finished scraping %v\n", channelID)
}

func GetChannelList(Discord *discordgo.Session, serverID string) {
	/*
	 * GetChannelList gets a list of (accessable given the users permissions)
	 * channels in a specified server.
	 */
	guild, err := Discord.Guild(serverID)
	channels, err := Discord.GuildChannels(serverID)
	if err != nil {
		Fatal(err.Error())
	}
	/*
	 * A tabwriter instance to prettify printing the important information about
	 * every channel we can find.
	 */
	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug|tabwriter.AlignRight)
	fmt.Fprintf(writer, "Channel list for %v (%v)\n", guild.Name, guild.ID)
	fmt.Fprintln(writer, "ID\tNAME\tTYPE\tNSFW\t")
	for _, channel := range channels {
		fmt.Fprintln(writer, strings.Replace(fmt.Sprintf("%v\t %v\t %v\t %t",
			channel.ID, channel.Name, ChannelType(channel.Type),
			channel.NSFW), "\"", "", -1))
	}
	writer.Flush()
}

func GetGuildDetails(Discord *discordgo.Session, serverID string) {
	/*
	 * GetGuildDetails gets useful information about a guild, including
	 * the name, id, owners id, region, and roles.
	 */
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
	 * Go through every role and print the information that matters about them,
	 * cleanly.
	 */
	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug|tabwriter.AlignRight)
	fmt.Fprintln(writer, "ROLE\tMANAGED\tHOISTED\tCOLOR\tPOSITION")
	for _, role := range guild.Roles {
		fmt.Fprintln(writer,
			strings.Replace(
				fmt.Sprintf("%v\t %t\t %v\t %v\t %v",
					role.Name, role.Managed, role.Hoist, role.Color, role.Position),
				"\"", "", -1))
	}
	writer.Flush()
}

func ChannelType(channelType discordgo.ChannelType) string {
	/*
	 * ChannelType returns the name of a channel from discordgo's numerical ID.
	 */
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

func AllChannels(Discord *discordgo.Session, serverID string,
	channelType int) []string {
	/*
	 * AllChannels is called by GetLogs and GetImages to get a list of all the
	 * channels in the server.
	 */
	var allChannels []string
	channels, err := Discord.GuildChannels(serverID)
	if err != nil {
		Fatal(err.Error())
	}
	for _, channel := range channels {
		if int(channel.Type) == channelType {
			allChannels = append(allChannels, channel.ID)
		}
	}
	return allChannels
}
