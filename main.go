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

type Single struct {
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

type Users struct {
	Users []Last `json:`
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

func main() {
	fmt.Println("Super Advanced HFT BOT:")
	m := 199
	profit := 0.0
	portfolio := 0.0
	cash := 10000.0
	isTrade := false
	short := false
	long := false
	for {
		moving := movingAVG(getLast(), m)
		last, _ := strconv.ParseFloat(getSpread().Last, 64)
		prev, _ := strconv.ParseFloat(getLast()[1].Rate, 64)
		if moving < last && !isTrade && last > prev {
			fmt.Println("ENTER SHORT AT: " + getSpread().HighestBid)
			last, _ := strconv.ParseFloat(getSpread().HighestBid, 64)
			portfolio = last
			short = true
			isTrade = true
		} else if last < moving && !isTrade && last < prev {
			fmt.Println("ENTER BUY AT: " + getSpread().LowestAsk)
			last, _ := strconv.ParseFloat(getSpread().LowestAsk, 64)
			portfolio = last
			cash = cash - last
			long = true
			isTrade = true
		} else if (last < moving || portfolio-last > 10) && isTrade && long {
			fmt.Println("EXIT SELL AT: " + getSpread().HighestBid)
			last, _ := strconv.ParseFloat(getSpread().HighestBid, 64)
			pr := last - portfolio
			cash = last + cash
			profit = cash - 10000
			long = false
			isTrade = false
			fmt.Printf("Profit/Loss from trade %f\n", pr)
			portfolio = 0
			fmt.Printf("Cash: %f  Profit: $%f \n", cash, profit)
		} else if (last < moving || last-portfolio > 10) && isTrade && short {
			fmt.Println("EXIT BUY AT: " + getSpread().LowestAsk)
			last, _ := strconv.ParseFloat(getSpread().LowestAsk, 64)
			pr := portfolio - last
			portfolio = portfolio - last
			cash = cash + portfolio
			profit = cash - 10000
			short = false
			isTrade = false
			fmt.Printf("Profit/Loss from trade %f\n", pr)
			portfolio = 0
			fmt.Printf("Cash: %f  Profit: $%f \n", cash, profit)
		}
		time.Sleep(10000000000)

	}

}

func movingAVG(l []Last, mov int) float64 {
	avg := 0.0
	for x := 0; x < mov; x++ {
		c, _ := strconv.ParseFloat(l[x].Rate, 64)
		avg = avg + c
	}
	avg = avg / float64(mov)
	return avg
}

func getLast() []Last {
	url := "https://poloniex.com/public?command=returnTradeHistory&currencyPair=USDT_BTC"
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
	//	results := Last{}
	err1 := json.Unmarshal([]byte((body)), &results)
	if err1 != nil {
		//fmt.Println(err)
	}
	return results

}

func getSpread() Single {
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
	err1 := json.Unmarshal(textBytes, &people1)

	if err1 != nil {
		//fmt.Println(err)
	}
	return people1["USDT_BTC"]

}
