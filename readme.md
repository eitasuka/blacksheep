# Blacksheep: The Discord Tooling Powerhouse.

## Design

Blacksheep is built in Go, a fast and powerful compiled language, instead of the
more common JavaScript or Python based solutions. It primarily targets Linux
and macOS, but it is very likely builds will be Windows-compatible.

Blacksheep is designed to be:
* Completely free and open-source replacement for the closed-source and
overpriced (very likely malware-ridden) tools available from sketchy sellers;
* Faster and easier to use than current closed and open source alternatives;
* A way to bring together many of the fragmented tools available into a concise
package.

## Features

### Channel Scraper

Scrape a channel, or an entire server, for images, text logs, or
both.

See: `blacksheep help scrape`

### Bot Account Control

* Send messages

See: `blacksheep help control`

### Configuration

* A human-readable serialisation language (TOML)

See: Configuration

### Selfbot

* Convert text to regional indicators/emojis

* Random Copypastas (with which you can add your own)

* Custom commands

* Avatar Grabber

* Grab User Account Information

* OwOify

See: `blacksheep help self`

## Performance

### Channel Scraper

On my system, Blacksheep scrapes about **270 messages per second**. It is worth
noting that each channel in a server has its own thread, so this number doesn't
increase exponentially based on the amount of channels being scraped.

## Build Instructions

Before installing Blacksheep, make sure `go` is installed via your distro's
package manager, or for Windows, download it from `https://golang.org/dl/`.

If you're on Windows and want to build from source, make sure `git` is installed
by downloading it from `https://git-scm.com/`

If `$GOBIN` isn't already in your `$PATH`, you may need to add it. On Linux, you
can simply execute `export PATH="$PATH:<your home directory>/go/bin"`.

### Automagically


1. Install Blacksheep with `go`

    `go install github.com/t1ra/blacksheep`

2. Run BlackSheep from your terminal emulator

    `blacksheep --help`

### From source

1. Clone the repo


* For the stable master branch:

    `git clone https://github.com/t1ra/blacksheep/master`

* For the probably broken working branch:

    `git clone https://github.com/t1ra/blacksheep/working`


2. Install

    `go install`

3. Run Blacksheep from your terminal emulator

    `blacksheep --help`

## Configuration

Currently, a blacksheep.toml file can be located in two separate places,
depending on the target platform. If you're using Windows, it's  at
`%USER%/blacksheep.toml`. If you're using macOS or Linux, it's located in
`~/.config/blacksheep/blacksheep.toml`. When you first install Blacksheep, this
file wont exist -- one will be generated for you.

After executing Blacksheep for the first time and creating a `blacksheep.toml`
file, you can modify the following values:

#### `Token` : string

Your Discord user token, used to access many features in Blacksheep. You need to have this for
BlackSheep to work.

Example: `Token = "SUPER-SECRET-TOKEN"`

#### `SaveDirectory` : string

An absolute path to where you would like Blacksheep to save files. If you want Blacksheep to save
files relative to the current working directory, you can leave this blank.

Example: `SaveDirectory = "/mount/discordlogs"`

#### `SelfBotPrefix` : string

The prefix that the selfbot will look for. If you want to use the default prefix, `::`, you can
leave this blank.

#### `SelfBotCopypastas` : []string

An array of custom copypastas you'd like to add to the `copypasta` command. If you don't want to add
any, you can leave this blank.

Example: `Copypastas = ["something", "very", "funny"]`
