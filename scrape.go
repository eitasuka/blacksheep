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
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"mvdan.cc/xurls/v2"
)

// Scrape is called by main.go, it creates a directory for every channel being scraped. If more
// than one is being scraped, it utilises a workgroup to create a new thread for each channel.
func Scrape(Session *discordgo.Session, server, channel, directory string) {
	var wg sync.WaitGroup
	_, err := os.Stat(directory)
	if os.IsNotExist(err) || err == nil {
		os.Mkdir(directory, 0700)
	} else {
		Fatal("Failed to create server directory: " + err.Error())
	}
	if channel == "" {
		for _, c := range AllChannels(Session, server) {
			wg.Add(1)
			go ScrapeChannel(Session, directory, server, c, &wg)
		}
	} else {
		ScrapeChannel(Session, directory, server, channel, &wg)
	}
	wg.Wait()
}

// ScrapeChannel gets every message in a channel and saves it, downloading any media it finds.
func ScrapeChannel(Session *discordgo.Session, directory, server, channel string,
	wg *sync.WaitGroup) {
	defer wg.Done()
	var mediaWaitGroup sync.WaitGroup
	messages, err := Session.ChannelMessages(channel, 100, "", "", "")
	if err != nil {
		Warning(fmt.Sprintf("Failed to get messages in %v: %v", channel, err.Error()))
		return
	}
	var lastMessage *discordgo.Message
	err = os.Remove(directory + channel + ".log")
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		Fatal("Failed to delete previous log file: " + err.Error())
	}
	file, err := os.OpenFile(directory+channel+".log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		Fatal("Failed to create log file: " + err.Error())
	}
	_, err = os.Stat(directory + channel)
	if os.IsNotExist(err) || err == nil {
		os.Mkdir(directory+channel, 0700)
	} else {
		Fatal("Failed to create media directory: " + err.Error())
	}
	defer file.Close()
	fmt.Println("Scraping channel", channel)
	for len(messages) != 0 {
		for _, message := range messages {
			for _, user := range message.Mentions {
				message.Content = strings.NewReplacer(
					"<@"+user.ID+">", "@"+user.Username+"#"+user.Discriminator,
					"<@!"+user.ID+">", "@"+user.Username+"#"+user.Discriminator,
				).Replace(message.Content)
			}
			file.WriteString(fmt.Sprintf("%v\n%v\n\n", message.Author, message.Content))
			for _, url := range xurls.Strict().FindAllString(message.Content, -1) {
				mediaWaitGroup.Add(1)
				go DownloadMedia(url, server, directory+channel+"/", &mediaWaitGroup)
			}
			for _, attachment := range message.Attachments {
				mediaWaitGroup.Add(1)
				go DownloadMedia(attachment.URL, server, directory+channel+"/", &mediaWaitGroup)
			}
			for _, embed := range message.Embeds {
				if embed.Image != nil {
					mediaWaitGroup.Add(1)
					go DownloadMedia(embed.Image.URL, server, directory+channel+"/", &mediaWaitGroup)
				}
			}
			mediaWaitGroup.Wait()
		}
		lastMessage = messages[len(messages)-1]
		messages, err = Session.ChannelMessages(channel, 100, lastMessage.ID, "", "")
		if err != nil {
			Warning(fmt.Sprintf("Failed to get messages in %v: %v", channel, err.Error()))
			return
		}
	}
	Success("Finished scraping channel " + channel)
}

// DownloadMedia is used by ScrapeChannel to offload media downloading, and asynchronously download
// it.
func DownloadMedia(url, server, directory string, wg *sync.WaitGroup) {
	defer wg.Done()
	response, err := http.Get(url)
	if err != nil {
		Warning("Failed to download " + url + ": " + err.Error())
	}
	defer response.Body.Close()
	file, err := os.Create(directory + strings.TrimSuffix(filepath.Base(url),
		filepath.Ext(url)) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(url))
	if err != nil {
		Fatal("Failed to create file: " + err.Error())
	}
	defer file.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		Warning("Failed to write image: " + err.Error())
	}
}

// AllChannels finds every channel it has access to, and makes sure its a guild text chat.
// (Voice channels, and even categories count as "channels")
func AllChannels(Session *discordgo.Session, server string) []string {
	var allChannels []string
	channels, err := Session.GuildChannels(server)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			Fatal("No --server provided to scrape.")
		}
		Fatal(err.Error())
	}
	for _, channel := range channels {
		if int(channel.Type) == 0 { /* Type 0 is a guild text chat */
			allChannels = append(allChannels, channel.ID)
		}
	}
	return allChannels
}
