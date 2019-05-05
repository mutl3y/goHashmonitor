package hashmonitor

import (
	"fmt"
	"time"
)

// var wg sync.WaitGroup

func Mine() {
	err := ConfigLogger("logging.conf", false)
	if err != nil {
		fmt.Printf("failed to configure logging")
	}

	cards := NewCardData(cfg)
	err = cards.GetStatus()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	debug(cards.String())
	if err = cards.ResetCards(false); err != nil {
		fmt.Printf("error Resetting cards %v\n", err)
	}

	api := NewStatsService(cfg).(*apiService)
	met := NewMetricsClient()
	if err = met.Config(cfg); err != nil {
		log.Infof("failed to config metrics client")
	}
	met.refresh = 10 * time.Second
	go met.backGroundWriter()

	go api.Monitor(met)

	m := NewMiner()
	ctx, err := m.ConfigMiner(cfg)
	if err != nil {
		log.Errorf("Failed configuring miner: %v\n", err)
	}

	err = m.StartMining(ctx)
	if err != nil {
		log.Errorf("failed to start mining %v", err)
	}
	// defer func(){
	//
	// }()

	go m.ConsoleMetrics(met)

	err = api.startingHash(300, 17*time.Second)
	if err != nil {
		log.Errorf("currenthash %v", err)
	}
	go api.showMonitor()
	err = api.currentHash(300, 2, 1*time.Second)
	if err != nil {
		log.Errorf("currenthash %v", err)
	}
	if err = m.StopMining(); err != nil {
		log.Errorf("failed to stop miner %v\n", err)
	}
	// miningTime := 1 * time.Minute
	// time.Sleep(miningTime)

	api.stopMonitor(met)
}

// func profitMine() {
// 	cards := NewCardData()
// 	err := cards.GetStatus(cfg)
// 	if err != nil {
// 		fmt.Printf("%+v\n", err)
// 	}
// }
