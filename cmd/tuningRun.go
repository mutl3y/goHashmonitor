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
	"time"
)

var EnableCommandSorting = false

// TuningCmd IntensityCmd represents the tuneIntensity command
var TuningCmd = &cobra.Command{
	Use:   "tuningRun",
	Short: "Tune intensity and worksize",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tuningRun called")
		c, err := hashmonitor.Config()
		if err != nil {
			return
		}

		var (
			is, ie, ii, ws, we, wi int
		)
		flags := cmd.Flags()
		i, err := flags.GetIntSlice("intensity")
		w, err := flags.GetIntSlice("worksize")

		is = i[0]
		ie = i[1]
		ii = i[2]
		if len(w) < 3 {
			log.Fatal("if you specify a worksize you must specify start, stop and increment, To stop worksize changing set start and stop to the same and set increment to 1")
		}
		ws = w[0]
		we = w[1]
		wi = w[2]

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

		autotune, err := flags.GetInt("autotune")
		if err != nil {

		}

		interleave, err := flags.GetInt("interleave")
		if err != nil {

		}
		afterLock, err := flags.GetDuration("afterLock")
		if err != nil {

		}
		runtime, err := flags.GetDuration("runtime")
		if err != nil {

		}

		session := hashmonitor.IntensityRun{
			Intensity: hashmonitor.IntRunArgs{Start: is, Stop: ie, Inc: ii},
			Worksize:  hashmonitor.IntRunArgs{Start: ws, Stop: we, Inc: wi},
			Runtime:   runtime,
			AutoTune:  autotune, Interleave: interleave, AfterAllLock: afterLock, ResetCards: reset}

		err = hashmonitor.TuningRun(c, session)

		fmt.Printf("%v", err)
	},
}

func init() {
	rootCmd.AddCommand(TuningCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ses := &hashmonitor.IntensityRun{}

	TuningCmd.Flags().String("influxIP", "", "influx, not needed if correct in config file")
	TuningCmd.Flags().Int("influxPort", 0, "influx port, not needed if correct in config file")
	TuningCmd.Flags().String("influxDB", "", "influx DB, not needed if correct in config file")
	TuningCmd.Flags().IntP("interleave", "i", 40, "interleave")
	TuningCmd.Flags().BoolP("reset", "r", false, "reset cards before each run")
	TuningCmd.Flags().IntP("autotune", "A", 0, "autotune, default off")
	TuningCmd.Flags().DurationP("afterLock", "L", 1*time.Minute, "if autotuning, how long to continue after threads are locked")
	TuningCmd.Flags().DurationP("runtime", "R", 60*time.Minute, "runtime ie 10m for ten minutes or 1h for 1 hour")
	TuningCmd.Flags().IntSliceP("intensity", "I", []int{1000, 2000, 100}, "instensity: start,stop,increment")
	TuningCmd.Flags().IntSliceP("worksize", "W", []int{4, 18, 2}, "worksize: start,stop,increment")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// IntensityCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
