// +build !windows

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

import "os/user"

var (
	/*
	 * This seems a little hacky, but its probably the best way to expand
	 * the user's home directory.
	 */
	usr, _          = user.Current()
	configDirectory = usr.HomeDir + "/.config/blacksheep/blacksheep.toml"
	configFolder    = usr.HomeDir + "/.config/blacksheep/"
)
