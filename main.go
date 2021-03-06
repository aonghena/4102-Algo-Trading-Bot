package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

/**
Structs used to handle API json data as well as Portfolios for trading
*/
//poleniex spread data
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

//last price data
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

//Currency Spread data
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

//User Portfolio Info
type Portfolio struct {
	TradingPair string
	Cash        float64
	Profit      float64
	Position    float64
	InitalCash  float64
	IsShort     bool
	IsLong      bool
}

/**
Calculates the moving average by using the trade history and the period (mov)
*/
func movingAVG(l []Last, mov int) float64 {
	avg := 0.0
	for x := 0; x < mov; x++ {
		c, _ := strconv.ParseFloat(l[x].Rate, 64)
		avg = avg + c
	}
	avg = avg / float64(mov)
	return avg
}

/**
Gets historical prices of the trading pair from the API
*/
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

/**
Gets the current price and spread from the API
*/
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
	//one line if/else to error check
	if err1 := json.Unmarshal(textBytes, &people1); err1 != nil {
		fmt.Println(err)
	}
	return people1[s]

}

/**
Function to conduct trades
*
Trades are based of off moving average for period as well as previous price.
*
*/
func trade(p Portfolio, mov int) Portfolio {
	moving := movingAVG(getLast(p.TradingPair), mov)
	spreadInfo := getSpread(p.TradingPair)
	last, _ := strconv.ParseFloat(spreadInfo.Last, 64)
	prev, _ := strconv.ParseFloat(getLast(p.TradingPair)[1].Rate, 64)
	//OPEN SHORT
	//
	if moving < last && !p.IsLong && !p.IsShort && last > prev {
		fmt.Println(p.TradingPair)
		last, _ := strconv.ParseFloat(spreadInfo.HighestBid, 64)
		fmt.Printf("ENTER SHORT AT: $%.3f | %s \n", last, time.Now().Format("2006-01-02 15:04:05"))
		p.Position = last
		p.IsShort = true
	} else if last < moving && !p.IsLong && !p.IsShort && last < prev {
		fmt.Println(p.TradingPair)
		last, _ := strconv.ParseFloat(spreadInfo.LowestAsk, 64)
		fmt.Printf("ENTER BUY AT: $%.3f | %s \n", last, time.Now().Format("2006-01-02 15:04:05"))
		p.Position = last
		p.Cash = p.Cash - last
		p.IsLong = true
	} else if (last < moving || p.Position-last > 10) && p.IsLong && !p.IsShort {
		fmt.Println(p.TradingPair)
		last, _ := strconv.ParseFloat(spreadInfo.HighestBid, 64)
		fmt.Printf("EXIT SELL AT: $%.3f | %s \n", last, time.Now().Format("2006-01-02 15:04:05"))
		pr := last - p.Position
		p.Cash = last + p.Cash
		p.Profit = p.Cash - p.InitalCash
		go tradeConfirmation(p, pr)
		p.IsLong = false

	} else if (last < moving || last-p.Position > 10) && !p.IsLong && p.IsShort {
		fmt.Println(p.TradingPair)
		last, _ := strconv.ParseFloat(spreadInfo.LowestAsk, 64)
		fmt.Printf("EXIT BUY AT: $%.3f | %s \n", last, time.Now().Format("2006-01-02 15:04:05"))
		pr := p.Position - last
		p.Position = p.Position - last
		p.Cash = p.Cash + p.Position
		p.Profit = p.Cash - p.InitalCash
		go tradeConfirmation(p, pr)
		p.IsShort = false
	}
	return p
}

/**
Trade Confimation output which is run concurrent
*/
func tradeConfirmation(p Portfolio, pr float64) {
	fmt.Printf("Profit/Loss from trade %f\n", pr)
	fmt.Printf("Cash: %f  Profit: $%f \n", p.Cash, p.Profit)
}

/**
Initilizes portfolio vales
*/
func createPortfolio(tradingPair string, cash float64) Portfolio {
	return Portfolio{tradingPair, cash, 0.0, 0.0, cash, false, false}
}

func main() {
	fmt.Println("Trading BOT:")
	fmt.Println("Status: Online")
	btc := createPortfolio("USDT_BTC", 10000.00)
	eth := createPortfolio("USDT_ETH", 1000.00)
	fmt.Printf("Currently Trading: %s, %s\n", btc.TradingPair, eth.TradingPair)
	//using the last 199 trades as moving average
	m := 199
	for {
		btc = trade(btc, m)
		eth = trade(eth, m)
		time.Sleep(10000000000) //pauses in order to stay inline with API call limitations
	}

}
