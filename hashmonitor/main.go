package hashmonitor

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

// var wg sync.WaitGroup

func Mine(c *viper.Viper) {
	err := ConfigLogger("logging.conf", false)
	if err != nil {
		fmt.Printf("failed to configure logging")
	}

	cards := NewCardData(c)
	err = cards.GetStatus()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	debug(cards.String())
	if err = cards.ResetCards(false); err != nil {
		fmt.Printf("error Resetting cards %v\n", err)
	}

	api := NewStatsService(c).(*apiService)
	met := NewMetricsClient()
	met.enabled = true
	met.refresh = 10 * time.Second
	met.db = "testMine"

	if err = met.Config(c); err != nil {
		log.Infof("failed to config metrics client")
	}

	go met.backGroundWriter()

	go api.Monitor(met)

	m := NewMiner()
	ctx, err := m.ConfigMiner(c)
	if err != nil {
		log.Errorf("Failed configuring miner: %v\n", err)
	}

	err = m.StartMining(ctx)
	if err != nil {
		log.Errorf("failed to start mining %v", err)
	}
	defer func(m *miner) {
		if err = m.StopMining(); err != nil {
			log.Errorf("failed to stop miner %v\n", err)
		}
	}(m)

	go m.ConsoleMetrics(met)

	err = api.startingHash(200, 17*time.Second)
	if err != nil {
		log.Errorf("currenthash %v", err)
	}
	go api.showMonitor()
	err = api.currentHash(300, 2, 1*time.Second)
	if err != nil {
		log.Errorf("currenthash %v", err)
	}
	api.stopMonitor(met)

	// miningTime := 1 * time.Minute
	// time.Sleep(miningTime)

}

// func profitMine() {
// 	cards := NewCardData()
// 	err := cards.GetStatus(cfg)
// 	if err != nil {
// 		fmt.Printf("%+v\n", err)
// 	}
// }
