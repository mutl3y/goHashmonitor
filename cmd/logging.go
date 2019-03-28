// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
)

// loggingCmd represents the logging command
var loggingCmd = &cobra.Command{
	Use:   "logging",
	Short: "use this to reset logging config to defaults",
	Long:  `you can either delete or use this to reset logging to defaults`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Flags().BoolP("force", "f", false, "you need to force to overwrite exisiting file")
		fmt.Println("logging called")
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			fmt.Printf("failed %v", err)
		}
		err = hashmonitor.ConfigLogger("logging.conf", force)
		if err != nil {
			fmt.Printf("failed %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(loggingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loggingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loggingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
