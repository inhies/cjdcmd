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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/inhies/go-cjdns/admin"
	"github.com/inhies/go-cjdns/key"
	"github.com/spf13/cobra"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/user"
	"regexp"
	"strings"
)

// gotYes will read from stdin and if it is any variation of 'y' or 'yes' then it returns true
// If defaultYes is set to true and the user presses enter without entering anything else it returns true
func gotYes(defaultYes bool) bool {
	var choice string
	n, _ := fmt.Scanln(&choice)
	if n == 0 {
		if defaultYes {
			return true
		} else {
			return false
		}
	}
	if strings.ToLower(choice) == "y" || strings.ToLower(choice) == "yes" {
		return true
	}
	return false
}

// Reads the .cjdnsadmin file and returns the structured contents
func readCjdnsadmin(file string) (admin *admin.CjdnsAdminConfig, err error) {
	rawFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	raw, err := stripComments(rawFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(raw, &admin)
	if err != nil {
		return nil, err
	}
	if err != nil {
		// BUG(inhies): Find a better way of dealing with these errors.
		if e, ok := err.(*json.SyntaxError); ok {
			// BUG(inhies): Instead of printing x amount of characters, print the previous and following 2 lines
			fmt.Println("Invalid JSON") //" at byte", e.Offset, "(after stripping comments...)")
			fmt.Println("----------------------------------------")
			fmt.Println(string(raw[e.Offset-60 : e.Offset+60]))
			fmt.Println("----------------------------------------")
		} else if _, ok := err.(*json.InvalidUTF8Error); ok {
			fmt.Println("Invalid UTF-8")
		} else if e, ok := err.(*json.InvalidUnmarshalError); ok {
			fmt.Println("Invalid unmarshall type", e.Type)
			fmt.Println(err)
		} else if e, ok := err.(*json.UnmarshalFieldError); ok {
			fmt.Println("Invalid unmarshall field", e.Field, e.Key, e.Type)
		} else if e, ok := err.(*json.UnmarshalTypeError); ok {
			fmt.Println("Invalid JSON")
			fmt.Println("Expected", e.Type, "but received a", e.Value)
			fmt.Println("I apologize for not being more helpful")
		} else if e, ok := err.(*json.UnsupportedTypeError); ok {
			fmt.Println("Invalid JSON")
			fmt.Println("I am unable to utilize type", e.Type)
		} else if e, ok := err.(*json.UnsupportedValueError); ok {
			fmt.Println("Invalid JSON")
			fmt.Println("I am unable to utilize value", e.Value, e.Str)
		}
		return nil, err
	}
	return

}

// Check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

// Attempt to read the .cjdnsadmin file from the users home directory
func loadCjdnsadmin() (cjdnsAdmin *admin.CjdnsAdminConfig, err error) {
	sudo_user := os.Getenv("SUDO_USER")
	var read_err error
	var tUser *user.User
	if sudo_user != "" {
		tUser, read_err = user.Lookup(sudo_user)
		if !fileExists(tUser.HomeDir + "/.cjdnsadmin") {
			tUser, read_err = user.Current()
		}
	} else {
		tUser, read_err = user.Current()
	}
	if read_err != nil {
		return
	}
	cjdnsAdmin, err = readCjdnsadmin(tUser.HomeDir + "/.cjdnsadmin")
	if err != nil {
		return
	}
	return
}

// Fills out an IPv6 address to the full 32 bytes
// This shouldn't be needed in newer versions of cjdns

func padIPv6(ip net.IP) string {
	raw := hex.EncodeToString(ip)
	parts := make([]string, len(raw)/4)
	for i := range parts {
		parts[i] = raw[i*4 : (i+1)*4]
	}
	return strings.Join(parts, ":")
}

/*
// Dumps the entire routing table and structures it
func getTable(user *cjdns.Conn) cjdns.Routes {
	response, err := user.NodeStore_dumpTable()
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}
	return response
}
*/
/*
		page := 0
		var more int64
		table = make([]*Route, 0)
		for more = 1; more != 0; page++ {
			response, err := user.NodeStore_dumpTable(page)
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			// If an error field exists, and we have an error, return it
			if _, ok := response["error"]; ok {
				if response["error"] != "none" {
					err = fmt.Errorf(response["error"].(string))
					fmt.Printf("Error: %v\n", err)
					return
				}
			}
			//Thanks again to SashaCrofter for the table parsing
			rawTable := response["routingTable"].([]interface{})
			for i := range rawTable {
				item := rawTable[i].(map[string]interface{})
				rPath := item["path"].(string)
				sPath := strings.Replace(rPath, ".", "", -1)
				bPath, err := hex.DecodeString(sPath)
				if err != nil || len(bPath) != 8 {
					//If we get an error, or the
					//path is not 64 bits, discard.
					//This should also prevent
					//runtime errors.
					continue
				}
				path := binary.BigEndian.Uint64(bPath)
				table = append(table, &Route{
					IP:      item["ip"].(string),
					RawPath: path,
					Path:    rPath,
					RawLink: item["link"].(int64),
					Link:    float64(item["link"].(int64)) / magicalLinkConstant,
					Version: item["version"].(int64),
				})

			}

			if response["more"] != nil {
				more = response["more"].(int64)
			} else {
				break
			}
		}
	return
*/

type Target struct {
	Target   string
	Supplied string
}

func validIP(input string) (result bool)   { return ipRegex.MatchString(input) }
func validPath(input string) (result bool) { return pathRegex.MatchString(input) }
func validHost(input string) (result bool) { return hostRegex.MatchString(input) }

// Sets target.Target to the requried IP or cjdns path
func setTarget(data []string, usePath bool) (target Target, err error) {
	if len(data) == 0 {
		err = fmt.Errorf("Invalid target specified")
		return
	}
	input := data[0]
	if input != "" {

		if validIP(input) {
			target.Supplied = data[0]
			target.Target = padIPv6(net.ParseIP(input))
			return

		} else if validPath(input) && usePath {
			target.Target = input
			target.Supplied = data[0]
			return

		} else if validHost(input) {
			var ips []string
			ips, err = resolveHost(input)
			if err != nil {
				return
			}
			// Return the first result
			for _, addr := range ips {
				target.Target = addr
				target.Supplied = input
				return
			}

		} else {
			err = fmt.Errorf("Invalid IPv6 address, cjdns path, or hostname")
			return
		}
	}

	if usePath {
		err = fmt.Errorf("You must specify an IPv6 address, hostname, or cjdns path")
		return
	}
	err = fmt.Errorf("You must specify an IPv6 address or hostname")
	return
}

// randString returns a random alphanumeric string where length is <= max >= min
func randString(min, max int) string {
	r := myRand(min, max, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	return r
}

// Returns a random character from the specified string where length is <= max >= min
func myRand(min, max int, char string) string {

	var length int

	if min < max {
		length = min + rand.Intn(max-min)
	} else {
		length = min
	}

	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = char[rand.Intn(len(char)-1)]
	}
	return string(buf)
}

func stripComments(b []byte) ([]byte, error) {
	regComment, err := regexp.Compile("(?s)//.*?\n|/\\*.*?\\*/")
	if err != nil {
		return nil, err
	}
	out := regComment.ReplaceAllLiteral(b, nil)
	return out, nil
}

func pubKeyToIPCmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		os.Exit(1)
	}

	for _, s := range args {
		k, err := key.DecodePublic(s)
		if err != nil {
			fmt.Println("Error converting key:", err)
			os.Exit(1)
		}

		fmt.Println(k.IP())
	}
}
