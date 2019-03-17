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
	"time"
)

var lower, upper int

// tuneInterleaveCmd represents the tuneInterleave command
var InterleaveCmd = &cobra.Command{
	Use:   "tuneInterleave",
	Short: "Tune interleave",
	Long: `This will backup the current gpu config file
it will then iterate over the range provided by upper --> lower 
performing a 60s benchmark that logs hashrate and interleave messages per second

You should find stable settings for intensity etc before attempting to tune interleave

If your miner is not stable for at least 5 minutes you should not run this...
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tuneInterleave called")
		cmd.Flags().IntP("lower", "l", 0, "Lowest interleave setting")
		cmd.Flags().IntP("upper", "u", 50, "Highest interleave setting")
		cmd.Flags().IntP("benchSec", "b", 60, "How long to run benchmark in seconds")
		lower, _ = cmd.Flags().GetInt("lower")
		upper, _ = cmd.Flags().GetInt("upper")
		fmt.Printf("%v %v %s", lower, upper, time.Now())
	},
}

func init() {
	rootCmd.AddCommand(InterleaveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tuneInterleaveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tuneInterleaveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
