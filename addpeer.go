package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/inhies/go-cjdns/config"
	"net"
	"os"
	"strings"
)

func addPassword(data []string) {
	// Load the config file
	if File == "" {
		cjdAdmin, err := loadCjdnsadmin()
		if err != nil {
			fmt.Println("Unable to load configuration file:", err)
			return
		}
		File = cjdAdmin.Config
		if File == "" {
			fmt.Println("Please specify the configuration file in your .cjdnsadmin file or pass the --file flag.")
			return
		}
	}
	fmt.Printf("Loading configuration from: %v... ", File)
	conf, err := config.LoadExtConfig(File)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	fmt.Printf("Loaded\n")

	var input string
	if len(data) == 0 {
		fmt.Printf("You didnt supply a password, should I generate one for you? [Y/n]: ")
		if gotYes(true) {
			for {
				input = randString(15, 50)
				fmt.Printf("Generated: '%v' Accept? [Y/n]: ", input)
				if gotYes(true) {
					break
				}
			}
		} else {
			// No password supplied, not going to generate one, I quit!
			return
		}
	} else {
		input = data[0]
	}

	if _, ok := conf["authorizedPasswords"]; !ok {
		conf["authorizedPasswords"] = make([]interface{}, 0)
		fmt.Println("Your configuration file does not contain an 'authorizedPasswords' section, so one was created for you")
	}
	passwords := conf["authorizedPasswords"].([]interface{})

	for loc, p := range passwords {
		x := p.(map[string]interface{})
		if x["password"] == input {
			fmt.Printf("Password '%v' exists with the following information:\n", input)
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
	pass["password"] = input

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

	// Get the permissions from the input file
	stats, err := os.Stat(File)
	if err != nil {
		fmt.Println("Error getting permissions for original file:", err)
		return
	}

	if File != "" && OutFile == "" {
		OutFile = File
	}

	// Check if the output file exists and prompt befoer overwriting
	if _, err := os.Stat(OutFile); err == nil {
		fmt.Printf("Overwrite %v? [y/N]: ", OutFile)
		if !gotYes(false) {
			return
		}
	}

	fmt.Printf("Saving configuration to: %v... ", OutFile)
	err = config.SaveConfig(OutFile, conf, stats.Mode())
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
func addPeer(data []string) {
	if len(data) == 0 {
		fmt.Println("You must enter the peering details surrounded by single qoutes '<peer details>'")
		return
	}
	input := data[0]

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

	// Load the config file
	if File == "" {
		cjdAdmin, err := loadCjdnsadmin()
		if err != nil {
			fmt.Println("Unable to load configuration file:", err)
			return
		}
		File = cjdAdmin.Config
		if File == "" {
			fmt.Println("Please specify the configuration file in your .cjdnsadmin file or pass the --file flag.")
			return
		}
	}
	fmt.Printf("Loading configuration from: %v... ", File)
	conf, err := config.LoadExtConfig(File)
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
	stats, err := os.Stat(File)
	if err != nil {
		fmt.Println("Error getting permissions for original file:", err)
		return
	}

	if File != "" && OutFile == "" {
		OutFile = File
	}

	// Check if the output file exists and prompt befoer overwriting
	if _, err := os.Stat(OutFile); err == nil {
		fmt.Printf("Overwrite %v? [y/N]: ", OutFile)
		if !gotYes(false) {
			return
		}
	}

	fmt.Printf("Saving configuration to: %v... ", OutFile)
	err = config.SaveConfig(OutFile, conf, stats.Mode())
	if err != nil {
		fmt.Println("Error saving config:", err)
		return
	}
	fmt.Printf("Saved\n")
}
