package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc/jsonrpc"
	"os"
	"strings"
	"strconv"
)

type Stock struct{
	StockName string
	StockPercent int
}

type StockRequest struct {
	Budget float32
	StockDetails []Stock
}

type TradeIdRequest struct{
	TradeId string
}

var stockReq=new(StockRequest)
var tradeIdRequest=new(TradeIdRequest)

func main() {
	defer func() {
			 if r := recover(); r != nil {
					 displayErrorMessage()
			 }
	 }()
	argsPassed:=os.Args
	var stocks []Stock
	if(len(argsPassed)==2){
		tradeIdRequest.TradeId=argsPassed[1]
		callRPC(1)
	}else if(len(argsPassed)==3){
		budget,_:=strconv.ParseFloat(argsPassed[1],32)
		stockReq.Budget=float32(budget)
		stocksDetails:= strings.Split(argsPassed[2],",")
		stocks = make([]Stock,len(stocksDetails))
		//countStockPercent:=0
		for i,values:= range stocksDetails{
				stock:=strings.Split(values,":")
				percent,_:=strconv.Atoi(strings.Trim(stock[1],"%"))
				//countStockPercent+=percent
				stockStruct:= Stock{stock[0],percent}
				stocks[i]=stockStruct
		}
		stockReq.StockDetails = stocks
		callRPC(2)
	}else{
		displayErrorMessage()
		return
	}
}


func callRPC(number int){
	client, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply string
	c := jsonrpc.NewClient(client)
	if number==2{
	err = c.Call("TradingServer.BuyStock", stockReq, &reply)
	}else{
		err = c.Call("TradingServer.CheckPortfolio", tradeIdRequest, &reply)
	}
	if err != nil {
		log.Fatal("Trading Server error:", err)
	}
	fmt.Print(reply)
}

func displayErrorMessage(){
  fmt.Println("Please enter valid input to buy stock or to check your portfolio")
  fmt.Println("To buy Stocks, please enter:  budget \"stock1:percentage,stock2:percentage\"")
  fmt.Println("To check your portfolio, please enter:  \"Trade ID\"")
}
