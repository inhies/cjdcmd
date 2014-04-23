/*
 * You may redistribute this program and/or modify it under the terms of
 * the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func listPasswordCmd(cmd *cobra.Command, args []string) {
	admin := Connect()
	p, err := admin.AuthorizedPasswords_list()
	if err != nil {
		fmt.Println("Error getting passwords from cjdns:", err)
		os.Exit(1)
	}

	for _, x := range p {
		fmt.Println(x)
	}
}
