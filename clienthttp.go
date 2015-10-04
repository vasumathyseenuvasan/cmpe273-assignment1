package main

import (
  "net/http"
  "fmt"
  "io/ioutil"
  "encoding/json"
  "bytes"
	"os"
	"strconv"
	"strings"
)

type BuyStockResponse struct{
  Type ResponseBuyStock `json:"result"`
  Start int `json:"error"`
  Count int `json:"id"`
}

type CheckPortfolioResponse struct{
  Type ResponseCheckPortfolio `json:"result"`
  Start int `json:"error"`
  Count int `json:"id"`
}

type ResponseCheckPortfolio struct{
	Stocks string  `json:"stocks"`
	CurrentMarketValue string `json:"currentMarketValue"`
	UnvestedAmount string `json:"unvestedAmount"`
}

type ClientRequest struct {
  		Method string         `json:"method"`
  		Params [1]interface{}     `json:"params"`
  		Id     uint64         `json:"id"`
}

type StockRequest struct {
	Budget float32
	StockDetails []Stock
}

type Stock struct{
	StockName string
	StockPercent int
}

type ResponseBuyStock struct{
	TradeId string `json:"tradeId"`
	Stocks string  `json:"stocks"`
	UnvestedAmount string `json:"unvestedAmount"`
}

type TradeIdRequest struct{
	TradeId string `json:"tradeId"`
}

var stockReq StockRequest

func main(){

defer func() {
 if r := recover(); r != nil {
	 displayErrorMessage()
 }
}()
argsPassed:=os.Args
var isCheckPortfolio bool
var stocks []Stock
var clientRequest ClientRequest
var tradeIdRequest TradeIdRequest
if(len(argsPassed)==2){
	isCheckPortfolio = true
	tradeIdRequest.TradeId=argsPassed[1]
	clientRequest.Method="TradingServer.CheckPortfolio"
	clientRequest.Params[0]=tradeIdRequest
	clientRequest.Id=0
}else if(len(argsPassed)==3){
	isCheckPortfolio = false
	budget,_:=strconv.ParseFloat(argsPassed[1],32)
	stockReq.Budget=float32(budget)
	stocksDetails:= strings.Split(argsPassed[2],",")
	stocks = make([]Stock,len(stocksDetails))
	for i,values:= range stocksDetails{
			stock:=strings.Split(values,":")
			percent,_:=strconv.Atoi(strings.Trim(stock[1],"%"))
			stockStruct:= Stock{stock[0],percent}
			stocks[i]=stockStruct
	}
	stockReq.StockDetails = stocks
	clientRequest.Method="TradingServer.BuyStock"
	clientRequest.Params[0]=stockReq
	clientRequest.Id=0
}else{
  displayErrorMessage()
  return
}

vari,_:=json.Marshal(clientRequest)
resp, _ := http.Post("http://localhost:8080/rpc", "application/json", bytes.NewBuffer(vari))
body, _ := ioutil.ReadAll(resp.Body)
if(!isCheckPortfolio){
v:=&BuyStockResponse{}
json.Unmarshal(body, &v)
defer resp.Body.Close()
fmt.Println("Trade Id: ",v.Type.TradeId)
fmt.Println("Stocks: \"",v.Type.Stocks,"\"")
fmt.Println("Unvested Amount:",v.Type.UnvestedAmount)
}else{
	v:=&CheckPortfolioResponse{}
	json.Unmarshal(body, &v)
	defer resp.Body.Close()
	fmt.Println("Stocks: \"",v.Type.Stocks,"\"")
	fmt.Println("Current Market Value: ",v.Type.CurrentMarketValue)
	fmt.Println("Unvested Amount:",v.Type.UnvestedAmount)
}
}

func displayErrorMessage(){
  fmt.Println("Please enter valid input to buy stock or to check your portfolio")
  fmt.Println("To buy Stocks, please enter:  budget \"stock1:percentage,stock2:percentage\"")
  fmt.Println("To check your portfolio, please enter:  \"Trade ID\"")
}
