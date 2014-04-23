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
	"github.com/inhies/go-cjdns/admin"
	"github.com/inhies/go-cjdns/config"
	"github.com/spf13/cobra"
	"os"
)

func cleanConfigCmd(cmd *cobra.Command, args []string) {
	var cjdnsAdmin *admin.CjdnsAdminConfig
	var err error

	if ConfFileIn == "" {
		if AdminFileIn == "" {
			cjdnsAdmin, err = loadCjdnsadmin()
			if err != nil {
				fmt.Println("Unable to load configuration file:", err)
				return
			}
		} else {
			cjdnsAdmin, err = readCjdnsadmin(AdminFileIn)
			if err != nil {
				fmt.Println("Error loading cjdnsadmin file:", err)
				return
			}
		}

		if cjdnsAdmin.Config == "" {
			fmt.Println("Please specify the configuration file in your .cjdnsadmin file or pass the --file flag.")
			os.Exit(1)
		}

		ConfFileIn = cjdnsAdmin.Config
	}

	fmt.Printf("Loading configuration from: %v... ", ConfFileIn)
	conf, err := config.LoadExtConfig(ConfFileIn)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	fmt.Printf("Loaded\n")

	// Get the permissions from the input file
	stats, err := os.Stat(ConfFileIn)
	if err != nil {
		fmt.Println("Error getting permissions for original file:", err)
		return
	}

	if ConfFileIn != "" && ConfFileOut == "" {
		ConfFileOut = ConfFileIn
	}

	// Check if the output file exists and prompt befoer overwriting
	if _, err := os.Stat(ConfFileOut); err == nil {
		fmt.Printf("Overwrite %v? [y/N]: ", ConfFileOut)
		if !gotYes(false) {
			return
		}
	}

	fmt.Printf("Saving configuration to: %v... ", ConfFileOut)
	err = config.SaveConfig(ConfFileOut, conf, stats.Mode())
	if err != nil {
		fmt.Println("Error saving config:", err)
		os.Exit(1)
	}
	fmt.Printf("Saved\n")
}
