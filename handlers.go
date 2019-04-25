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

	"github.com/logrusorgru/aurora"
)

// Fatal is called when something is critically wrong, like a vital config file not existing, or
// not having the right permissions to create something.
func Fatal(err string) {
	fmt.Println(aurora.Red(aurora.Bold("Error:")), aurora.Red(err))
	Warning("If this sounds like a bug in BlackSheep, it can be reported via a pull request to" +
		" github.com/t1ra/blacksheep.")
	os.Exit(1)
}

// Warning is called when something should be *noted* but isn't exactly Fatal.
func Warning(err string) {
	fmt.Println(aurora.Brown(aurora.Bold("Warning:")), aurora.Brown(err))
}

// Success is used to note something of significance.
func Success(str string) {
	fmt.Println(aurora.Green(str))
}
