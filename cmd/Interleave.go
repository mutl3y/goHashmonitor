// Copyright © 2019 Mark Heynes hashmonitor@heynes.biz
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"goHashmonitor/hashmonitor"
	"time"

	"github.com/spf13/cobra"
)

var lower, upper int

// InterleaveCmd represents the tuneInterleave command
var InterleaveCmd = &cobra.Command{
	Use:   "Interleave",
	Short: "Tune interleave",
	Long: `This will mine for given time cyling through the provided interleave range
uses intensity and worksize from amd.conf soo make sure these are good too start with
review results in Grafana to find the perfect value

If your miner is not stable for at least 5 minutes you should not run this...`,
	Run: func(cmd *cobra.Command, args []string) {

		c, err := hashmonitor.Config()
		if err != nil {
			return
		}

		var (
			is, ie, ii int
		)
		flags := cmd.Flags()
		i, err := flags.GetIntSlice("interleave")

		is = i[0]
		ie = i[1]
		ii = i[2]

		c.Set("Influx.Enabled", true)

		err = c.BindPFlag("Influx.IP", cmd.Flags().Lookup("influxIP"))
		if err != nil {

		}

		err = c.BindPFlag("Influx.Port", cmd.Flags().Lookup("influxPort"))
		if err != nil {

		}

		err = c.BindPFlag("Influx.DB", cmd.Flags().Lookup("influxDB"))
		if err != nil {

		}
		err = c.BindPFlag("Core.Stak.Dir", cmd.Flags().Lookup("stakdirectory"))
		if err != nil {

		}
		reset, err := flags.GetBool("reset")
		if err != nil {

		}

		logging, err := flags.GetBool("logging")
		if err != nil {

		}
		if logging {
			fmt.Println("enabling logging")
			err := hashmonitor.ConfigLogger("logging.conf", false)
			if err != nil {
				fmt.Println("failed to configure logging %v", err)
			}
		}

		runtime, err := flags.GetDuration("runtime")
		if err != nil {

		}

		session := hashmonitor.InterleaveRun{
			Interleave: hashmonitor.IntRunArgs{Start: is, Stop: ie, Inc: ii},
			Runtime:    runtime,
			ResetCards: reset}

		err = hashmonitor.InterleaveSession(c, session)

		fmt.Printf("%v", err)
	},
}

func init() {
	rootCmd.AddCommand(InterleaveCmd)

	InterleaveCmd.Flags().String("influxIP", "", "influx, not needed if correct in config file")
	InterleaveCmd.Flags().Int("influxPort", 0, "influx port, not needed if correct in config file")
	InterleaveCmd.Flags().String("influxDB", "", "influx DB, not needed if correct in config file")
	InterleaveCmd.Flags().BoolP("reset", "r", false, "reset cards before each run")
	InterleaveCmd.Flags().DurationP("runtime", "R", 2*time.Minute, "runtime ie 10m for ten minutes or 1h for 1 hour")
	InterleaveCmd.Flags().IntSliceP("interleave", "I", []int{20, 40, 1}, "interleave: start,stop,increment")
	InterleaveCmd.Flags().StringP("stakdirectory", "D", "xmr-stak", "xmr-stak folder, not needed if specified in config file")
	InterleaveCmd.Flags().BoolP("logging", "L", false, "enable logging to file / Slack etc logging.conf")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tuneInterleaveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tuneInterleaveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
