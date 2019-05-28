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

// IntensityCmd represents the tuneIntensity command
var IntensityCmd = &cobra.Command{
	Use:   "tuneIntensity",
	Short: "Tune intensity for card",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tuneIntensity called")
		c, err := hashmonitor.Config()
		if err != nil {
			return
		}
		t := hashmonitor.NewTestPlan("test")
		err = t.Run(c)
		if err != nil {
			log.Fatalf("failed to run testplan %v\n %v", t.Name, err)
		}

		fmt.Printf("%v", c)
	},
}

func init() {
	// 	rootCmd.AddCommand(IntensityCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// IntensityCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// IntensityCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
