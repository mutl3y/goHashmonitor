package hashmonitor

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

var pathSep = string(os.PathSeparator)
var root string
var cfg, _ = Config()

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
			if err := c.WriteConfig(); err != nil {
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

	cfg := viper.New()
	cfg.SetConfigFile("config.yaml")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath(root)

	cfg.SetDefault("Core.Connection.Check.Seconds", 10)
	cfg.SetDefault("Core.Connection.Check.Destination", "www.google.co.uk")
	cfg.SetDefault("Core.Debug", false)
	cfg.SetDefault("Core.Display.Destination", "Local")
	cfg.SetDefault("Core.Display.Port", 8080)
	cfg.SetDefault("Core.Log.Configfile", "logging.conf")
	// cfg.SetDefault("Core.Log.File", "hashmonitor.log")
	// cfg.SetDefault("Core.Log.Rotate", true)
	// cfg.SetDefault("Core.Log.Rotate_Hours", 24)
	cfg.SetDefault("Core.Reboot.Enabled", false)
	cfg.SetDefault("Core.Reboot.Timeout_Seconds", 15)

	cfg.SetDefault("Core.Stak.Args", []string{"--noNVIDIA"})
	cfg.SetDefault("Core.Stak.Dir", root+"xmr-stak")
	cfg.SetDefault("Core.Stak.Exe", "./xmr-stak.exe")
	cfg.SetDefault("Core.Stak.Ip", "192.168.0.28")
	cfg.SetDefault("Core.Stak.Port", 420)
	cfg.SetDefault("Core.Stak.Refresh_MS", time.Millisecond*500)
	cfg.SetDefault("Core.Stak.Timeout", 2)
	cfg.SetDefault("Core.Stak.Start_Attempts", 3)
	cfg.SetDefault("Core.Stak.Stable_Secs", 60)
	cfg.SetDefault("Core.Stak.Seconds_Before_Rate_check", 20)
	cfg.SetDefault("Core.Stak.Tools", []string{
		"OverdriveNTool.exe -consoleonly -r1 -p1XMR",
		"OverdriveNTool.exe -consoleonly -r2 -p2XMR",
	})

	cfg.SetDefault("Device.Hash_Drop", 300)
	cfg.SetDefault("Device.Reset.Enabled", false)
	cfg.SetDefault("Device.Reset.OnStartup", false)
	cfg.SetDefault("Device.Reset.Interval", 3)
	cfg.SetDefault("Device.MaxResetSecs", 3)
	cfg.SetDefault("Device.Count", 1)

	cfg.SetDefault("Influx.Enabled", false)
	cfg.SetDefault("Influx.IP", "192.168.0.29")
	cfg.SetDefault("influx.DB", "hashmonitor")
	cfg.SetDefault("Influx.Port", 8089)
	cfg.SetDefault("Influx.FlushSec", 10*time.Second)

	cfg.SetDefault("Slack.Url", "https://hooks.slack.com/services/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")
	cfg.SetDefault("Slack.Username", "unsetUserName")
	cfg.SetDefault("Slack.PeriodicReporting.Enabled", false)
	cfg.SetDefault("Slack.PeriodicReporting.Minutes", 10)
	cfg.SetDefault("Slack.MessageWindow", time.Duration(time.Second*30))
	cfg.SetDefault("Slack.Verbosity", 2)

	cfg.SetDefault("Profit.Display", false)
	cfg.SetDefault("Profit.Refresh_Time", 20)
	cfg.SetDefault("Profit.StatsHour", 1)
	cfg.SetDefault("Profit.CoinStats", false)
	cfg.SetDefault("Profit.Calc.Heavy", 1.1)
	cfg.SetDefault("Profit.Calc.v2", 0.95)
	cfg.SetDefault("Profit.Calc.Fast", 1.8)
	cfg.SetDefault("Profit.Calc.Lite", 2.5)
	cfg.SetDefault("Profit.Switching.enabled", false)
	cfg.SetDefault("Profit.Switching.Live", false)
	cfg.SetDefault("Profit.Switching.KillRunning", false)
	cfg.SetDefault("Profit.Switching.Percentage", 5)
	cfg.SetDefault("Profit.Switching.CheckMins", 20)

	cfg.SetDefault("Temperature.Enabled", false)
	cfg.SetDefault("Temperature.MaxC", 34)
	cfg.SetDefault("Temperature.MinDiff", 0.2)
	cfg.SetDefault("Temperature.Valid_Minutes", 10)
	cfg.SetDefault("Temperature.Active", false)
	cfg.SetDefault("Temperature.CSV", "./sensor-data")
	cfg.SetDefault("Temperature.Column", "InnerTemp")
	cfg.SetDefault("Temperature.Stop_Mining", false)

	cfg.SetDefault("Tuning.Benchmark.Enable", false)
	cfg.SetDefault("Tuning.Benchmark.RunSecs", false)
	cfg.SetDefault("Tuning.Intensity.Analyser", false)
	cfg.SetDefault("Tuning.Intensity.Stable_Minutes", 3)
	cfg.SetDefault("Tuning.Intensity.Precision", 3)

	return cfg
}
