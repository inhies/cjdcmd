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
	"encoding/json"
	"fmt"
	"github.com/inhies/go-cjdns/admin"
	cjdnsConfig "github.com/inhies/go-cjdns/config"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
)

func cjdnsAdminCmd(cmd *cobra.Command, args []string) {
	var cjdnsAdmin *admin.CjdnsAdminConfig
	var err error

	if ConfFileIn == "" {
		if AdminFileIn == "" {
			cjdnsAdmin, err = loadCjdnsadmin()
			if err != nil {
				fmt.Println("Unable to load configuration file:", err)
				os.Exit(1)
			}
		} else {
			cjdnsAdmin, err = readCjdnsadmin(AdminFileIn)
			if err != nil {
				fmt.Println("Error loading cjdnsadmin file:", err)
				os.Exit(1)
			}
		}

		if cjdnsAdmin.Config == "" {
			fmt.Println("Please specify the configuration file in your .cjdnsadmin file or pass the --file flag.")
			os.Exit(1)
		}

		ConfFileIn = cjdnsAdmin.Config
	}

	fmt.Printf("Loading configuration from: %v... ", ConfFileIn)
	conf, err := cjdnsConfig.LoadMinConfig(ConfFileIn)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	fmt.Printf("Loaded\n")

	split := strings.LastIndex(conf.Admin.Bind, ":")
	addr := conf.Admin.Bind[:split]
	port := conf.Admin.Bind[split+1:]
	portInt, err := strconv.Atoi(port)
	if err != nil {
		fmt.Println("Error with cjdns admin bind settings")
		return
	}

	adminOut := admin.CjdnsAdminConfig{
		Addr:     addr,
		Port:     portInt,
		Password: conf.Admin.Password,
		Config:   ConfFileIn,
	}

	jsonout, err := json.MarshalIndent(adminOut, "", "\t")
	if err != nil {
		fmt.Println("Unable to create JSON for .cjdnsadmin")
		return
	}

	if AdminFileOut == "" {
		tUser, err := user.Current()
		if err != nil {
			fmt.Println("I was unable to get your home directory, please manually specify where to save the file with --outfile")
			return
		}
		AdminFileOut = tUser.HomeDir + "/.cjdnsadmin"
	}

	// Check if the output file exists and prompt befoer overwriting
	if _, err := os.Stat(AdminFileOut); err == nil {
		fmt.Printf("Overwrite %v? [y/N]: ", AdminFileOut)
		if !gotYes(false) {
			return
		}
	} else {
		fmt.Println("Saving to", AdminFileOut)
	}

	ioutil.WriteFile(AdminFileOut, jsonout, 0600)
}
