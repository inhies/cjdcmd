package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func passGenCmd(cmd *cobra.Command, args []string) {
	if len(args) > 0 && len(args[0]) > 0 {
		fmt.Println(args[0] + "_" + randString(25, 50))
	} else {
		fmt.Println(randString(25, 50))
	}
}
