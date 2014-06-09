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
	"bufio"
	"encoding/json"
	"fmt"
	cjdnsConfig "github.com/inhies/go-cjdns/config"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func init() {
	AddPeerCmd.Flags().StringVarP(&ConfFileIn, "file", "f", "",
		"the cjdroute.conf configuration file to")
	AddPeerCmd.Flags().StringVarP(&ConfFileOut, "outfile", "o", "",
		"the configuration file to save to")
}

func addPeerCmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		os.Exit(1)
	}

	if ConfFileIn == "" {
		cjdnsAdmin, err := loadCjdnsadmin()
		if err != nil {
			fmt.Fprintln(os.Stderr, "cjdroute.conf not specified with '-f' and could not read cjdnsadmin file")
			os.Exit(1)
		}

		if cjdnsAdmin.Config == "" {
			fmt.Println("Please specify the configuration file with --file or in ~/.cjdnsadmin.")
			os.Exit(1)
		}

		ConfFileIn = cjdnsAdmin.Config
	}

	for _, input := range args {

		// Strip comments, just in case, and surround with {} to make it valid JSON
		raw, err := stripComments([]byte("{" + input + "}"))
		if err != nil {
			fmt.Println("Comment errors: ", err)
			return
		}

		// Convert from JSON to an object
		var object map[string]interface{}
		err = json.Unmarshal(raw, &object)
		if err != nil {
			fmt.Println("JSON Error:", err)
			return
		}

		fmt.Printf("Loading configuration from: %v... ", ConfFileIn)
		conf, err := cjdnsConfig.LoadExtConfig(ConfFileIn)
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}
		fmt.Printf("Loaded\n")

		if _, ok := conf["interfaces"]; !ok {
			fmt.Println("Your configuration file does not contain an 'interfaces' section")
			return
		}

		is := conf["interfaces"].(map[string]interface{})
		if len(is) == 0 {
			fmt.Println("No valid interfaces found!")
			return
		} else if len(is) > 1 {
			fmt.Println("You have multiple interfaces to choose from, enter yes or no, or press enter for the default option:")
		}

		var useIface string
		var i []interface{}

	selectIF:
		for {
			for key, _ := range is {
				if len(is) > 1 {
					fmt.Printf("Add peer to '%v' [Y/n]: ", key)
					if gotYes(true) {
						i = is[key].([]interface{})
						useIface = key
						break selectIF
					}
				} else if len(is) == 1 {
					i = is[key].([]interface{})
					useIface = key
				}
			}
			if useIface == "" {
				fmt.Println("You must select an interface to add to!")
				continue
			}
			break
		}
		var iX map[string]interface{}
		if len(i) > 1 {
			fmt.Printf("You have multiple '%v' options to choose from, enter yes or no, or press enter for the default option\n", useIface)
		selectIF2:
			for _, iFace := range i {
				temp := iFace.(map[string]interface{})
				fmt.Printf("Add peer to '%v %v' [Y/n]: ", useIface, temp["bind"])
				if gotYes(true) {
					iX = iFace.(map[string]interface{})
					break
				}
			}
			if iX == nil {
				fmt.Println("You must select an interface to add to!")
				goto selectIF2
			}
		} else if len(i) == 1 {
			iX = i[0].(map[string]interface{})
		} else {
			fmt.Printf("No valid settings for '%v' found!\n", useIface)
			return
		}

		peers := iX["connectTo"].(map[string]interface{})

		for key, data := range object {
			var peer map[string]interface{}
			if peers[key] != nil {
				peer = peers[key].(map[string]interface{})
				fmt.Printf("Peer '%v' exists with the following information:\n", key)
				for f, v := range peer {
					fmt.Printf("\t\"%v\":\"%v\"\n", f, v)
				}

				fmt.Printf("Update peer with new information? [Y/n]: ")
				if gotYes(true) {
					peer = data.(map[string]interface{})
					fmt.Printf("Updating peer '%v'\n", key)
				} else {
					fmt.Printf("Skipped updating peer '%v'\n", key)
					continue
				}
			} else {
				fmt.Printf("Adding new peer '%v'\n", key)
				peer = data.(map[string]interface{})
			}

			// Optionally add meta information
			for {
				r := bufio.NewReader(os.Stdin)
				fmt.Printf("Enter a field name for any extra information, or press enter to skip: ")
				fName, _ := r.ReadString('\n')
				fName = strings.TrimSpace(fName)
				if len(fName) == 0 {
					break
				}

				fmt.Printf("Enter a the content for field '%v' or press enter to cancel: ", fName)
				fData, _ := r.ReadString('\n')
				fData = strings.TrimSpace(fData)
				if len(fData) == 0 {
					continue
				}

				peer[fName] = fData
				continue
			}

			fmt.Println("Peer information:")

			for f, v := range peer {
				fmt.Printf("\t\"%v\":\"%v\"\n", f, v)
			}

			fmt.Printf("Add this peer? [Y/n]: ")
			if gotYes(true) {
				peers[key] = peer
				fmt.Println("Peer added")
			} else {
				fmt.Println("Skipped adding peer")
			}

		}

		// Get the permissions from the input file
		stats, err := os.Stat(ConfFileIn)
		if err != nil {
			fmt.Println("Error getting permissions for original file:", err)
			return
		}

		if ConfFileOut == "" {
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
		err = cjdnsConfig.SaveConfig(ConfFileOut, conf, stats.Mode())
		if err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
		fmt.Printf("Saved\n")
	}
}
