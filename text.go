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
 * text.go implements a few functions for printing pretty errors, these are used
 * all across other files.
 */
package main

import (
	"fmt"
	"os"

	"github.com/logrusorgru/aurora"
)

func Fatal(err string) {
	fmt.Println(aurora.Red(aurora.Bold("Error:")), aurora.Red(err))
	Warning("This may be a bug with BlackSheep, and should be reported via a" +
		" pull request to github.com/t1ra/BlackSheep.")
	os.Exit(1)
}

func Warning(err string) {
	fmt.Println(aurora.Brown(aurora.Bold("Warning:")), aurora.Brown(err))
}

func Notice(str string) {
	fmt.Println(aurora.Gray(str))
}

func Success(str string) {
	fmt.Println(aurora.Green(str))
}
