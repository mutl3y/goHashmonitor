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
	"github.com/spf13/cobra"
	"goHashmonitor/hashmonitor"
	"log"
	"os"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Generates config file",
	Long:    `Used to generate configuration file`,
	Example: `hashmonitor config -force`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := hashmonitor.DefaultConfig()
		f := cfg.ConfigFileUsed()
		if cmd.Flag("force").Changed {
			if err := cfg.WriteConfig(); err != nil {
				log.Fatalf("Error writing config File: %v", err)
			}
			log.Fatalf("Created %v Check contents", f)
		}

		if _, err := os.Stat(f); os.IsNotExist(err) {
			if err := cfg.WriteConfig(); err != nil {
				log.Fatalf("Error saving %v, %v", f, err)
			}
		} else {
			log.Fatalf("File exists, %v use --force to overwrite", f)

		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().BoolP("force", "f", false, "Overwrite an existing config")
}
