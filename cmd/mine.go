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
	"github.com/spf13/cobra"
	"goHashmonitor/hashmonitor"
	"log"
)

// mineCmd represents the mine command
var mineCmd = &cobra.Command{
	Use:   "mine",
	Short: "Start a standard mining session",
	Long: `
This will start a mining session ignoring any profit mining features
hashrate drop and restart options are still valid
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("standard mining mode")
		c, err := hashmonitor.Config()
		if err != nil {
			fmt.Println("config file issue", err)

			return
		}

		fname := c.GetString("Core.Log.Configfile")
		err = hashmonitor.ConfigLogger(fname, false)
		if err != nil {
			fmt.Println("issue configuring logging", err)
			return
		}

		flags := rootCmd.Flags()
		hashmonitor.Debug, err = flags.GetBool("debugOutput")
		if err != nil {
			fmt.Printf("error setting debug flag %v", err)
		}

		err = c.BindPFlag("Core.Stak.Dir", cmd.Flags().Lookup("stakdirectory"))
		if err != nil {
			fmt.Printf("unable to set stak directory in config %v", err)
		}

		ms, err := hashmonitor.NewMineSession(c)
		if err != nil {
			log.Fatalf("%v", err)
		}

		ms.Api.Monitor(ms.Met)
		defer ms.Api.StopMonitor(ms.Met)
		err = ms.Mine()
		if err != nil {
			fmt.Printf("error mining %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mineCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mineCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mineCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
