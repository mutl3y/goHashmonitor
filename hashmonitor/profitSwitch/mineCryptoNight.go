package profitSwitch

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"time"
)

type reward struct {
	TickerSymbol      string
	Algorithm         string
	LastNetworkUpdate time.Time
	LastPriceUpdate   time.Time
	Btc               float64
}

type rewards []reward

func (r rewards) Len() int {
	return len(r)
}

func (r rewards) Less(i, j int) bool {
	return r[i].Btc < r[j].Btc
}

func (r rewards) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r reward) String() (str string) {
	str += fmt.Sprintf("%-6v\t", r.TickerSymbol)
	str += fmt.Sprintf("%-20v\t", r.Algorithm)
	str += fmt.Sprintf("BTC %0.20f\t", r.Btc)
	str += fmt.Sprintf("Last Update %v", r.LastPriceUpdate.Format("2006-01-02 15:04:05"))
	return
}

type CoinStats struct {
	// 	*sync.RWMutex
	lastUpdate time.Time
	data       []rewards
}

type MineCryptoNight interface {
	Stats()
}

func GetMineCryptoNightStats(hr, limit int) (rewards, map[string]float64, error) {
	coinStats := make(rewards, 0, 100)

	type coins struct {
		HashRate                int     `json:"hash_rate"`
		CryptonightV2V4Factor   float64 `json:"cryptonight-v2_v4_factor"`
		CryptonightFastV2Factor float64 `json:"cryptonight-fast-v2_factor"`
		CryptonightHeavyFactor  float64 `json:"cryptonight-heavy_factor"`
		CryptonightLiteFactor   float64 `json:"cryptonight-lite_factor"`
		Rewards                 []struct {
			TickerSymbol      string    `json:"ticker_symbol"`
			Algorithm         string    `json:"algorithm"`
			LastNetworkUpdate time.Time `json:"last_network_update"`
			LastPriceUpdate   time.Time `json:"last_price_update"`
			Reward24H         struct {
				Coins float64 `json:"coins"`
				Btc   float64 `json:"btc"`
				Usd   float64 `json:"usd"`
				Eur   float64 `json:"eur"`
			} `json:"reward_24h"`
			Reward7D struct {
				Coins float64 `json:"coins"`
				Btc   float64 `json:"btc"`
				Usd   float64 `json:"usd"`
				Eur   float64 `json:"eur"`
			} `json:"reward_7d"`
			Reward30D struct {
				Coins float64 `json:"coins"`
				Btc   float64 `json:"btc"`
				Usd   float64 `json:"usd"`
				Eur   float64 `json:"eur"`
			} `json:"reward_30d"`
			Reward1Y struct {
				Coins float64 `json:"coins"`
				Btc   float64 `json:"btc"`
				Usd   float64 `json:"usd"`
				Eur   float64 `json:"eur"`
			} `json:"reward_1y"`
		} `json:"rewards"`
	}

	url := fmt.Sprintf("https://minecryptonight.net/api/rewards?hr=%v&limit=%v", hr, limit)
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	// timeoutError := "request canceled"

	res, err := client.Get(url)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading body %v", err)
	}
	out := &coins{}
	err = json.Unmarshal(body, &out)
	if err != nil {
		fmt.Println(err)

	}
	err = res.Body.Close()
	if err != nil {
		log.Fatalf("failed to close body, stopping under the no leak policy %v", err)
	}

	for _, v := range out.Rewards {
		r := reward{}
		r.Algorithm = v.Algorithm
		r.LastNetworkUpdate = v.LastNetworkUpdate
		r.LastPriceUpdate = v.LastPriceUpdate
		r.TickerSymbol = v.TickerSymbol
		r.Btc = v.Reward24H.Btc

		coinStats = append(coinStats, r)
	}
	sort.Sort(sort.Reverse(coinStats))
	con := make(map[string]float64)
	con["CryptonightFastV2Factor "] = out.CryptonightFastV2Factor
	con["CryptonightHeavyFactor "] = out.CryptonightHeavyFactor
	con["CryptonightLiteFactor "] = out.CryptonightLiteFactor
	con["CryptonightV2V4Factor "] = out.CryptonightV2V4Factor
	return coinStats, con, err
}

type customCalc map[string]float64

func NewCustomCalcs(c *viper.Viper) customCalc {
	cu := make(customCalc)
	cu["cryptonight-conceal"] = c.GetFloat64("Profit.Calc.Conceal")
	cu["cryptonight-v2"] = c.GetFloat64("Profit.Calc.v2")
	cu["cryptonight-v4"] = c.GetFloat64("Profit.Calc.v4")
	cu["cryptonight-heavy"] = c.GetFloat64("Profit.Calc.Heavy")
	cu["cryptonight-heavy-x"] = c.GetFloat64("Profit.Calc.HeavyX")
	cu["cryptonight-haven"] = c.GetFloat64("Profit.Calc.Haven")
	cu["cryptonight-saber"] = c.GetFloat64("Profit.Calc.Saber")
	cu["cryptonight-lite-v1"] = c.GetFloat64("Profit.Calc.Litev1")
	cu["cryptonight-fast-v2"] = c.GetFloat64("Profit.Calc.Fastv2")

	return cu
}

func (va customCalc) custCalc(algo string, con map[string]float64) (algoFamily string, multi float64, err error) {
	switch a := algo; a {
	case "cryptonight":
		return "Cryptonight", 1.0, nil
	case "cryptonight-v2":
		fallthrough
	case "cryptonight-v4":
		algoFamily = "CryptonightV2V4Factor"

	case "cryptonight-heavy-x":
		fallthrough
	case "cryptonight-heavy":
		fallthrough
	case "cryptonight-haven":
		fallthrough
	case "cryptonight-saber":
		algoFamily = "CryptonightHeavyFactor"
	case "cryptonight-lite-v1":
		algoFamily = "CryptonightLiteFactor"
	case "cryptonight-conceal":
		fallthrough
	case "cryptonight-fast-v2":
		algoFamily = "CryptonightFastV2Factor"
	default:
		return "unknown", 1, fmt.Errorf("unsupported Algorithm, please report via issues: %v\n", algo)
	}

	multi = va[algo]

	return algoFamily, multi, nil
}

func (r *rewards) CustomCalcs(cu customCalc, con map[string]float64) error {
	fpPrecision := 100000000000000000000.0
	// 	fmt.Printf("%2.30f\n",fpPrecision)
	rp := *r
	for i, v := range rp {
		algofam, multi, err := cu.custCalc(v.Algorithm, con)
		if err != nil {
			return err
		}

		div := con[algofam]
		if div == 0 {
			div = 1
		}

		std := math.Round(v.Btc * fpPrecision)

		result := ((std / div) * multi) / fpPrecision
		rp[i].Btc = result
		if Debug {
			fmt.Printf("%20v %6v (%0.30f / %v ) * %v = %0.30f %0.30f %v\n", v.Algorithm, v.TickerSymbol, v.Btc, div, multi, result, rp[i].Btc, algofam)
		}

	}

	return nil
}

/*
switch( $coin.algorithm ){
						{$_-in"cryptonight-v2"}     {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-v2_factor' ) )*$cryptonightv2_factor ),
							                                                        10 ) )
						}
						{$_-in"cryptonight-fast"}     {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-fast_factor' ) )*$cryptonightfast_factor ),
							                                                        10 ) )
						}
						{$_-in"cryptonight-heavy", "cryptonight-saber", 'cryptonight-haven', 'cryptonight-webchain'}     {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-heavy_factor' ) )*$cryptonightheavy_factor ),
							                                                        10 ) )
						}
						{$_-in"cryptonight-lite", "cryptonight-lite-v1"}     {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-lite_factor' ) )*$cryptonightlite_factor ),
							                                                        10 ) )
						}
						Default {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( $coin.reward_24h.btc, 10 ) )
						}
*/
