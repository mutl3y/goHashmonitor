package hashmonitor

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/viper"
)

var pathSep = string(os.PathSeparator)
var root string

// var cfg, _ = Config()

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	root = cwd + pathSep
	// cfg, err = Config()
	// if err != nil {
	// 	fmt.Printf("failed setting globaL config %v",err)
	// }
}

func Config() (*viper.Viper, error) {
	c := DefaultConfig()
	f := c.ConfigFileUsed()
	err := c.MergeInConfig()
	if err != nil {
		if _, err = os.Stat(c.ConfigFileUsed()); os.IsNotExist(err) {
			log.Info("Config File Not Found")
			if err = c.WriteConfig(); err != nil {
				log.Fatalf("Error creating, %v", err)
			}
			log.Fatalf("Created %v, check contents", f)
		}
		return nil, fmt.Errorf("Error reading File %v\n", f)
	}
	err = c.WriteConfigAs(c.ConfigFileUsed())
	return c, err
}

func DefaultConfig() *viper.Viper {

	c := viper.New()
	c.SetConfigFile("config.yaml")
	c.SetConfigType("yaml")
	c.AddConfigPath(root)
	switch Os := runtime.GOOS; {
	case Os == "windows":
		c.SetDefault("Core.Stak.Port", 420)
	case Os == "linux":
		c.SetDefault("Core.Stak.Port", 1420)
	default:
		log.Fatalf("Config() OS not supported")
	}

	c.SetDefault("Core.Connection.Check.Seconds", 10)
	c.SetDefault("Core.Connection.Check.Destination", "www.google.co.uk")
	c.SetDefault("Core.Debug", false)
	c.SetDefault("Core.Display.Destination", "Local")
	c.SetDefault("Core.Display.Port", 8080)
	c.SetDefault("Core.Hash.Drop", 50)
	c.SetDefault("Core.Hash.Min", 200)
	c.SetDefault("Core.Log.Configfile", "logging.conf")
	// c.SetDefault("Core.Log.File", "hashmonitor.log")
	// c.SetDefault("Core.Log.Rotate", true)
	// c.SetDefault("Core.Log.Rotate_Hours", 24)
	c.SetDefault("Core.Reboot.Enabled", false)
	c.SetDefault("Core.Reboot.Timeout_Seconds", 15)

	c.SetDefault("Core.Stak.Args", []string{"--noNVIDIA"})
	c.SetDefault("Core.Stak.Dir", root+"xmr-stak")
	c.SetDefault("Core.Stak.Exe", "./xmr-stak.exe")
	c.SetDefault("Core.Stak.Ip", "127.0.0.1")

	c.SetDefault("Core.Stak.Refresh_Time", time.Millisecond*500)
	c.SetDefault("Core.Stak.Timeout", 2)
	c.SetDefault("Core.Stak.Start_Attempts", 3)
	c.SetDefault("Core.Stak.Stable_Time", 30*time.Second)
	c.SetDefault("Core.Stak.Seconds_Before_Rate_check", 20)
	c.SetDefault("Core.Stak.Tools", []string{
		"OverdriveNTool.exe -consoleonly -r1 -p1XMR",
		"OverdriveNTool.exe -consoleonly -r2 -p2XMR",
	})

	c.SetDefault("Device.Reset.Enabled", false)
	c.SetDefault("Device.Reset.OnStartup", false)
	c.SetDefault("Device.Reset.Interval", 3)
	c.SetDefault("Device.MaxResetSecs", 3)
	c.SetDefault("Device.Count", 1)

	c.SetDefault("Influx.Enabled", false)
	c.SetDefault("Influx.IP", "192.168.0.29")
	c.SetDefault("influx.DB", "goHashmonitor")
	c.SetDefault("Influx.Port", 8086)
	c.SetDefault("Influx.User", nil)
	c.SetDefault("Influx.Pw", nil)
	c.SetDefault("Influx.FlushSec", 10*time.Second)

	c.SetDefault("Slack.Url", "https://hooks.slack.com/services/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")
	c.SetDefault("Slack.Username", "unsetUserName")
	c.SetDefault("Slack.PeriodicReporting.Enabled", false)
	c.SetDefault("Slack.PeriodicReporting.Minutes", 10)
	c.SetDefault("Slack.MessageWindow", time.Duration(time.Second*30))
	c.SetDefault("Slack.Verbosity", 2)

	c.SetDefault("Profit.Display", false)
	c.SetDefault("Profit.Refresh_Time", 20)
	c.SetDefault("Profit.StatsHour", 1)
	c.SetDefault("Profit.CoinStats", false)
	c.SetDefault("Profit.Calc.Heavy", 1.1)
	c.SetDefault("Profit.Calc.v2", 0.95)
	c.SetDefault("Profit.Calc.Fast", 1.8)
	c.SetDefault("Profit.Calc.Lite", 2.5)
	c.SetDefault("Profit.Switching.enabled", false)
	c.SetDefault("Profit.Switching.Live", false)
	c.SetDefault("Profit.Switching.KillRunning", false)
	c.SetDefault("Profit.Switching.Percentage", 5)
	c.SetDefault("Profit.Switching.CheckMins", 20)

	c.SetDefault("Temperature.Enabled", false)
	c.SetDefault("Temperature.MaxC", 34)
	c.SetDefault("Temperature.MinDiff", 0.2)
	c.SetDefault("Temperature.Valid_Minutes", 10)
	c.SetDefault("Temperature.Active", false)
	c.SetDefault("Temperature.CSV", "./sensor-data")
	c.SetDefault("Temperature.Column", "InnerTemp")
	c.SetDefault("Temperature.Stop_Mining", false)

	c.SetDefault("Tuning.Benchmark.Enable", false)
	c.SetDefault("Tuning.Benchmark.RunSecs", false)
	c.SetDefault("Tuning.Intensity.Analyser", false)
	c.SetDefault("Tuning.Intensity.Stable_Minutes", 3)
	c.SetDefault("Tuning.Intensity.Precision", 3)

	return c
}
