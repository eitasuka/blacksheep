# BlackSheep: [cool tagline]

## Design

BlackSheep is built in Go, a fast and powerful compiled language, instead of the
more common JavaScript-based solutions. It primarily targets Linux and macOS
systems, but Windows builds should generally be available.

BlackSheep is designed to be:
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

## Planned Features

### Selfbot

* Custom commands

* Copypastas

* Regional Indicator-ifier

* Message spam

### Token Generator

## Performance

### Channel Scraper

On my system, BlackSheep scrapes about 270 messages per second. It is worth
noting that each channel in a server has its own thread, so this number doesn't
increase exponentially based on the amount of channels being scraped.

## Build Instructions

Before installing BlackSheep, make sure `go` is installed via your distro's
package manager, or for Windows, download it from `https://golang.org/dl/`.

If you're on Windows and want to build from source, make sure `git` is installed
by downloading it from `https://git-scm.com/`

### Automagically


1. Install BlackSheep with `go`

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

3. Run BlackSheep from your terminal emulator

    `blacksheep --help`

## Configuration

Currently, a blacksheep.toml file can be located in two separate places,
depending on the target platform. If you're using Windows, it's placed wherever
BlackSheep is executed from (the current working directory,
`./blacksheep.toml`). If you're using macOS or Linux, it's located in
`~/.config/blacksheep.toml`. When you first install BlackSheep, this file wont
exist -- one will be generated for you.

After executing BlackSheep for the first time and creating a blacksheep.toml
file, you can modify the following values:

#### `Token` : String

Description: Your Discord user token, used to access many features in
BlackSheep.

Example: `Token = "SUPER-SECRET-TOKEN"`

#### `SaveDirectory` : String

Description: An absolute path to where you would like BlackSheep to save files.
If you want BlackSheep to save files relative to the current working directory,
you can leave this blank.

Example: `SaveDirectory = "/mount/discordlogs"`
