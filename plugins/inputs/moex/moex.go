package moex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Moex struct{
	start, limit int
	ac telegraf.Accumulator
	Log      telegraf.Logger `toml:"-"`
	Tickers []string `toml:"tickers"`
}

type History struct {
	Data [][]interface{} `json:"data"`
}

type HistoryResponse struct {
	History History `json:"history"`
}

func (m *Moex) Description() string {
	return ""
}

func (m *Moex) SampleConfig() string {
	return ""
}

func (m *Moex) Start(ac telegraf.Accumulator) error {
	m.ac = ac
	go func() {
		for _, ticker := range m.Tickers {
			start := 0
			limit := 100
			for {
				time.Sleep(time.Second * 5)
				url := fmt.Sprintf("http://iss.moex.com/iss/history/engines/%s/markets/%s/boards/TQBR/securities/%s.json?start=%d&limit=%d&history.columns=SECID,TRADEDATE,OPEN,HIGH,LOW,CLOSE", "stock", "shares", ticker, start, limit)
				m.Log.Info(url)
				resp, err := http.Get(url)
				if err != nil {
					fmt.Errorf("%s", err)
				}
				v, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Errorf("%s", err)
				}
		
				r := HistoryResponse{}
				err = json.Unmarshal(v, &r)
				if err != nil {
					fmt.Errorf("%s", err)
				}
				if len(r.History.Data) == 0 {
					m.Log.Info("end of data")
					break 
				}
				for _, t := range r.History.Data {
					tradedate, err := time.Parse("2006-01-02", t[1].(string))
					if err != nil {
						fmt.Errorf("%s", err)
					}
					openPrice, _ := t[2].(float64)
					m.ac.AddFields(
						"price", 
						map[string]interface{}{"value": openPrice},
						map[string]string{"type": "open", "ticker": ticker}, 
						tradedate,
					)
					highPrice, _ := t[3].(float64)
					m.ac.AddFields(
						"price", 
						map[string]interface{}{"value": highPrice},
						map[string]string{"type": "high", "ticker": ticker}, 
						tradedate,
					)
					lowPrice, _ := t[4].(float64)
					m.ac.AddFields(
						"price", 
						map[string]interface{}{"value": lowPrice},
						map[string]string{"type": "low", "ticker": ticker}, 
						tradedate,
					)
					closePrice, _ := t[5].(float64)
					m.ac.AddFields(
						"price", 
						map[string]interface{}{"value": closePrice},
						map[string]string{"type": "close", "ticker": ticker}, 
						tradedate,
					)
				}
				start = start + limit
			}
		}
	}()
	return nil
}

func (m *Moex) Stop() {
}

func (m *Moex) Gather(telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("moex", func() telegraf.Input {
		return &Moex{
			start: 0,
			limit: 100,
		}
	})
}