package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type PoloniexExchange map[string]struct {
	ID            int    `json:"id"`
	Last          string `json:"last"`
	LowestAsk     string `json:"lowestAsk"`
	HighestBid    string `json:"highestBid"`
	PercentChange string `json:"percentChange"`
	BaseVolume    string `json:"baseVolume"`
	QuoteVolume   string `json:"quoteVolume"`
	IsFrozen      string `json:"isFrozen"`
	High24Hr      string `json:"high24hr"`
	Low24Hr       string `json:"low24hr"`
}
type CurrencyPair struct {
	ID            int    `json:"id"`
	Last          string `json:"last"`
	LowestAsk     string `json:"lowestAsk"`
	HighestBid    string `json:"highestBid"`
	PercentChange string `json:"percentChange"`
	BaseVolume    string `json:"baseVolume"`
	QuoteVolume   string `json:"quoteVolume"`
	IsFrozen      string `json:"isFrozen"`
	High24Hr      string `json:"high24hr"`
	Low24Hr       string `json:"low24hr"`
}
type Last struct {
	GlobalTradeID int    `json:"globalTradeID"`
	TradeID       int    `json:"tradeID"`
	Date          string `json:"date"`
	Ttype         string `json:"type"`
	Rate          string `json:"rate"`
	Amount        string `json:"amount"`
	Total         string `json:"total"`
	OrderNumber   int    `json:"orderNumber"`
}

type Portfolio struct {
	TradingPair string
	Cash        float64
	Profit      float64
	Position    float64
	InitalCash  float64
	IsShort     bool
	IsLong      bool
}

var btc Portfolio

//Calculates moving average
func movingAVG(l []Last, mov int) float64 {
	avg := 0.0
	for x := 0; x < mov; x++ {
		c, _ := strconv.ParseFloat(l[x].Rate, 64)
		avg = avg + c
	}
	avg = avg / float64(mov)
	return avg
}

//Get Historic prices from the last x transactions to calculate moving average for algorithm
func getLast(pair string) []Last {
	url := "https://poloniex.com/public?command=returnTradeHistory&currencyPair=" + pair
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	var results []Last
	if err1 := json.Unmarshal([]byte((body)), &results); err1 != nil {
		fmt.Println(err)
	}
	return results

}

//Gets Current price Spread for currency pair
func getSpread(s string) CurrencyPair {
	url := "https://poloniex.com/public?command=returnTicker"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	textBytes := []byte((body))

	people1 := PoloniexExchange{}
	if err1 := json.Unmarshal(textBytes, &people1); err1 != nil {
		fmt.Println(err)
	}
	return people1[s]

}

func trade(p Portfolio, mov int) Portfolio {
	moving := movingAVG(getLast(p.TradingPair), mov)
	spreadInfo := getSpread(p.TradingPair)
	last, _ := strconv.ParseFloat(spreadInfo.Last, 64)
	prev, _ := strconv.ParseFloat(getLast(p.TradingPair)[1].Rate, 64)
	//OPEN SHORT
	//
	if moving < last && !p.IsLong && !p.IsShort && last > prev {
		fmt.Println(p.TradingPair)
		fmt.Println("ENTER SHORT AT: " + spreadInfo.HighestBid)
		last, _ := strconv.ParseFloat(spreadInfo.HighestBid, 64)
		p.Position = last
		p.IsShort = true
	} else if last < moving && !p.IsLong && !p.IsShort && last < prev {
		fmt.Println(p.TradingPair)
		fmt.Println("ENTER BUY AT: " + spreadInfo.LowestAsk)
		last, _ := strconv.ParseFloat(spreadInfo.LowestAsk, 64)
		p.Position = last
		p.Cash = p.Cash - last
		p.IsLong = true
	} else if (last < moving || p.Position-last > 10) && p.IsLong && !p.IsShort {
		fmt.Println(p.TradingPair)
		fmt.Println("EXIT SELL AT: " + spreadInfo.HighestBid)
		last, _ := strconv.ParseFloat(spreadInfo.HighestBid, 64)
		pr := last - p.Position
		p.Cash = last + p.Cash
		p.Profit = p.Cash - p.InitalCash
		go tradeConfirmation(p, pr)
		p.IsLong = false

	} else if (last < moving || last-p.Position > 10) && !p.IsLong && p.IsShort {
		fmt.Println(p.TradingPair)
		fmt.Println("EXIT BUY AT: " + spreadInfo.LowestAsk)
		last, _ := strconv.ParseFloat(spreadInfo.LowestAsk, 64)
		pr := p.Position - last
		p.Position = p.Position - last
		p.Cash = p.Cash + p.Position
		p.Profit = p.Cash - p.InitalCash
		go tradeConfirmation(p, pr)
		p.IsShort = false
	}
	return p
}

func tradeConfirmation(p Portfolio, pr float64) {
	if pr != -999 {
		fmt.Printf("Profit/Loss from trade %f\n", pr)
		fmt.Printf("Cash: %f  Profit: $%f \n", p.Cash, p.Profit)
	}

}

func createPortfolio(tradingPair string, cash float64) Portfolio {
	return Portfolio{tradingPair, cash, 0.0, 0.0, cash, false, false}
}

func main() {
	fmt.Println("Trading BOT:")
	fmt.Println("Status: Online")
	btc := createPortfolio("USDT_BTC", 10000.00)
	eth := createPortfolio("USDT_ETH", 1000.00)

	m := 199
	for {
		btc = trade(btc, m)
		eth = trade(eth, m)
		time.Sleep(10000000000)
	}

}
