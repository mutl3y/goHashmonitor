package hashmonitor

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type MineSession struct {
	confFile string
	api      *apiService
	ca       *CardData
	met      *metrics
	amdConf  AmdConf
}

func (s *MineSession) Mine(c *viper.Viper) error {
	gpuConf := s.amdConf.GpuThreadsConf
	if len(gpuConf) == 0 {
		return fmt.Errorf("no threads found in amd.txt")
	}

	err := s.ca.ResetCards(true)
	if err != nil {
		return fmt.Errorf("reset %v", err)
	}

	// config miner early so we can use kill method
	m := NewMiner()
	err = m.ConfigMiner(c)
	if err != nil {
		return fmt.Errorf("Failed configuring miner: %v\n", err)
	}

	err = m.killStak()
	if err != nil {
		log.Errorf("killStak %v", err)
	}

	// write amd.conf to influx
	tags := map[string]string{"type": "amdConf"}
	err = s.met.Write("config", tags, s.amdConf.Map())
	if err != nil {
		return fmt.Errorf("failed to write metrics %v", err)
	}
	err = s.met.Event(fmt.Sprintf("%+v", s.amdConf), "", "stak config")
	if err != nil {
		debug("failed to send event data %v", err)
	}

	// todo loop round current hash
	go s.api.showMonitor()
	defer s.api.stopMonitor(s.met)
	err = MiningSession(m, s.api, s.met)

	return nil
}

func MiningSession(m *miner, api *apiService, met *metrics) error {
	defer func() { _ = m.StopMining() }()
	err := m.StartMining()
	if err != nil {
		return fmt.Errorf("failed to start mining %v", err)
	}

	go m.ConsoleMetrics(met)

	err = api.startingHash(200, 20*time.Second, true)
	if err != nil {

	}
	err = api.currentHash(200, 5, 1*time.Second)
	if err != nil {
		return fmt.Errorf("currenthash %v", err)
	}

	return nil
}
