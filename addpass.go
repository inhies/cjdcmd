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
	"fmt"
	cjdnsConfig "github.com/inhies/go-cjdns/config"
	"github.com/spf13/cobra"
	"net"
	"os"
	"strings"
)

const safePasswordLength = 31

func addPasswordCmd(cmd *cobra.Command, args []string) {
	// Load the config file
	if Verbose {
		fmt.Printf("Loading %s...\t", ConfFileIn)
	}
	conf, err := cjdnsConfig.LoadExtConfig(ConfFileIn)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	if Verbose {
		fmt.Println("Loaded")
	}

	var password string
	if len(args) == 0 {
		fmt.Printf("You didnt supply a password, should I generate one for you? [Y/n]: ")
		if gotYes(true) {
			for {
				password = randString(15, 50)
				fmt.Printf("Generated: '%v' Accept? [Y/n]: ", password)
				if gotYes(true) {
					break
				}
			}
		} else {
			// No password supplied, not going to generate one, I quit!
			return
		}
	} else {
		password = args[0]
		l := len(password) + 1
		if l < safePasswordLength {
			fmt.Printf("Password looks short, pad? [Y/n]: ")
			if gotYes(true) {
				pad := randString(7, 25)
				password = fmt.Sprintf("%s-%s", password, pad)
			}
		}
	}

	if _, ok := conf["authorizedPasswords"]; !ok {
		conf["authorizedPasswords"] = make([]interface{}, 0)
		if Verbose {
			fmt.Println("Your configuration file does not contain an 'authorizedPasswords' section, so one was created for you")
		}
	}
	passwords := conf["authorizedPasswords"].([]interface{})

	for loc, p := range passwords {
		x := p.(map[string]interface{})
		if x["password"] == password {
			fmt.Printf("Password '%v' exists with the following information:\n", password)
			for f, v := range x {
				fmt.Printf("\t\"%v\":\"%v\"\n", f, v)
			}

			fmt.Printf("Update password with new information? [Y/n]: ")
			if gotYes(true) {
				// Remove the entry we are replacing
				passwords = append(passwords[:loc], passwords[loc+1:]...)
				break
			} else {
				return
			}
		}
	}

	pass := make(map[string]interface{})
	pass["password"] = password

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
				fmt.Printf("Add password to '%v' [Y/n]: ", key)
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

		pass[fName] = fData
		continue
	}

	fmt.Println("Password information:")

	for f, v := range pass {
		fmt.Printf("\t\"%v\":\"%v\"\n", f, v)
	}

	fmt.Printf("Add this password? [Y/n]: ")
	if gotYes(true) {
		conf["authorizedPasswords"] = append(passwords, pass)
		fmt.Println("Password added")
	} else {
		fmt.Println("Cancelled adding password")
		return
	}

	/*
		fmt.Printf("Load password into cjdns immediately? [Y/n]: ")
		c := Connect()
		if gotYes(true) {
			err = c.AuthorizedPasswords_add("", password, 0)
			// azazello is emery
			if err != nil {
				fmt.Printf("Failed to add password: %s\n", err)
			}
		}
	*/

	// Get the permissions from the password file
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
	err = cjdnsConfig.SaveConfig(ConfFileOut, conf, stats.Mode())
	if err != nil {
		fmt.Println("\nError saving config:", err)
		return
	}
	fmt.Printf("Saved\n")
	var bind string
	if strings.ToLower(useIface) == "ethinterface" {
		iFace, err := net.InterfaceByName(iX["bind"].(string))
		if err != nil {
			fmt.Println("Unable to get interface's MAC address, you'll have to enter it yourself")
			bind = "UNKNOWN"
		} else {
			bind = iFace.HardwareAddr.String()
		}
	} else {
		bind = iX["bind"].(string)
	}
	fmt.Println("Here are the details to be shared with your new peer:")
	fmt.Printf("\"%v\":{\n", bind)
	fmt.Printf("\t\"password\":\"%v\",\n", pass["password"].(string))
	fmt.Printf("\t\"publicKey\":\"%v\"\n", conf["publicKey"].(string))
	fmt.Printf("}\n")

}
