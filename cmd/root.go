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
	"os"
	"time"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goHashmonitor",
	Short: "XMR-STAK management tool",
	Long:  `Written to keep your miner hashing and taking care of card issues before they become a problem.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.MousetrapHelpText = "Command line tool, You have to provide parameters"
	rootCmd.PersistentFlags().BoolP("debugOutput", "d", false, "enable debugging output, disables on screen stats")
	rootCmd.PersistentFlags().DurationP("consolidate", "c", 10*time.Second, "consolidate debug messages by time period, eg: 1m2s")
	rootCmd.PersistentFlags().StringP("stakdirectory", "D", "xmr-stak", "xmr-stak folder, not needed if specified in config file")
	rootCmd.PersistentFlags().BoolP("stakOutput", "Z", false, "enable stak debugging output, very verbose")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// if cfgFile != "" {
	// 	// Use config file from the flag.
	// 	viper.SetConfigFile(cfgFile)
	// } else {
	// 	// Find home directory.
	// 	home, err := homedir.Dir()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(1)
	// 	}

	// 		// Search config in home directory with name ".goHashmonitor" (without extension).
	// 		viper.AddConfigPath(home)
	// 		viper.SetConfigName(".goHashmonitor")
	// 	}

	// viper.AutomaticEnv() // read in environment variables that match
	//
	// // If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Println("Using config file:", viper.ConfigFileUsed())
	// }

}

// func getHelp() {
// 	// 	fmt.Printf("%v",InterleaveCmd.HelpFunc())
// }
