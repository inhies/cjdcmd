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
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"time"
)

var (
	logLevel string
	logFile  string
	logLine  int
)

func init() {
	LogCmd.PersistentFlags().StringVarP(&logLevel, "level", "", "", "log level")
	LogCmd.PersistentFlags().StringVarP(&logFile, "file", "", "", "log level")
	LogCmd.PersistentFlags().IntVarP(&logLine, "line", "", -1, "log level")
}

const (
	format     = "%s %s %s:%d %s\n" // TODO: add user formatted output
	timeFormat = "15:04:05"
)

func logCmd(cmd *cobra.Command, args []string) {
	msgs := make(chan *admin.LogMessage)

	a := Connect()
	loggingStreamID, err := a.AdminLog_subscribe(logLevel, logFile, logLine, msgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	for {
		select {
		case m := <-msgs:
			//if !ok {
			//	fmt.Println("Error reading log response from cjdns.")
			//	os.Exit(1)
			//}
			fmt.Printf(format,
				time.Unix(m.Time, 0).Format(timeFormat),
				m.Level, m.File, m.Line, m.Message,
			)

		case <-sig:
			err = a.AdminLog_unsubscribe(loggingStreamID)
			if err != nil {
				fmt.Println("Error unsubscribing from log:", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}
