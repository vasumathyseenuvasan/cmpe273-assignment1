package main

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"strconv"
	"strings"
	"errors"
)


type StockRequest struct {
	Budget float32
	StockDetails []Stock
}

type Stock struct{
	StockName string
	StockPercent int
}

type StockBoughtDetails struct{
	unVestedAmount string
	stocksBought []StockBought
}

type StockBought struct{
	StockName string
	NumberOfStocks int
	BuyingPrice float32
}

type TradeIdRequest struct{
	TradeId string
}

var StockDetails = make(map[int]StockBoughtDetails)
var TradeID int = 0

type YahooResponse struct{
List struct{
    Meta struct{
      Type string `json:"type"`
      Start int `json:"start"`
      Count int `json:"count"`
    }
    Resources []struct{
      Resource struct{
        ClassName string `json:"classname"`
        Fields struct{
		        Change string `json:"change"`
		        Chg_Percent string `json:"chg_percent"`
		        Day_high string `json:"day_high"`
		        Day_low string `json:"day_low"`
		        Issuer_name string `json:"issuer_name"`
		        Issuer_name_lang string `json:"issuer_name_lang"`
		        Name string `json:"name"`
		        Price string `json:"price"`
		        Symbol string `json:"symbol"`
		        Ts string `json:"ts"`
		        Type string `json:"type"`
		        Utctime string `json:"utctime"`
		        Volume string `json:"volume"`
		        Year_high string `json:"year_high"`
		        Year_low string `json:"year_low"`
        }
      }
    }
  }
}

func main() {
	tradingServer := new(TradingServer)
	server := rpc.NewServer()
	server.Register(tradingServer)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			log.Print("new connection established\n")
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}
}

type TradingServer struct{}
func (this *TradingServer) BuyStock(args *StockRequest, reply *string) error {
	if(!checkPercentageStock(args)){
		return errors.New("Percentage of stocks should add upto 100")
	}
	*reply = callYahooServiceAPIBuyStock(args)
	return nil
}

func (this *TradingServer) CheckPortfolio(args *TradeIdRequest, reply *string) error {
	*reply = callYahooServiceAPICheckPortfolio(args)
	return nil
}

func checkPercentageStock(args *StockRequest) bool{
	countStockPercent:=0
	for _,stockDetails:=range args.StockDetails{
		countStockPercent+=stockDetails.StockPercent
		}
	if(countStockPercent!=100){
				return false
	}
	return true
}

// Call Yahoo Service API
func callYahooServiceAPIBuyStock(stockRequest *StockRequest) string{
	var stocks string
	for _,stockNames:=range stockRequest.StockDetails{
		stocks+=stockNames.StockName
		stocks+=","
	}
	v:= new(YahooResponse)
	v = callYahooServiceAPI(stocks)
	return calculateStocksBought(v,stockRequest)
}

func calculateStocksBought(response *YahooResponse,stockRequest *StockRequest) string{
	var stocksArray = make([]StockBought ,len(response.List.Resources))
	var stockbought StockBought
	var InVestedAmount float32
	var stocks string
	TradeID=TradeID+1
	for i,stocksResp:=range response.List.Resources{
		for _,stockDetails:=range stockRequest.StockDetails{
			if stocksResp.Resource.Fields.Symbol == stockDetails.StockName{
						stockbought.StockName=stockDetails.StockName
						amountToBuyStock:=stockRequest.Budget*(float32(stockDetails.StockPercent)/float32(100))
						stockPrice,_:=strconv.ParseFloat(stocksResp.Resource.Fields.Price,32)
						stockPrice32:=float32(stockPrice)
						numberOfStocksbought:=int(amountToBuyStock/stockPrice32)
						stockbought.NumberOfStocks=numberOfStocksbought
						amountUsedForStock:=float32(numberOfStocksbought)*stockPrice32
						stockbought.BuyingPrice=amountUsedForStock
						InVestedAmount+=amountUsedForStock
						stocks+=stockDetails.StockName+":"
						stocks+= strconv.Itoa(numberOfStocksbought)+":"
						stocks+=strconv.FormatFloat(float64(amountUsedForStock), 'f', 5, 32)+","
						stocksArray[i]=stockbought
						break
			}
		}
	}
	unvestedAmount:=stockRequest.Budget-InVestedAmount
	stocks=strings.TrimRight(stocks,",")
	var stockDetails StockBoughtDetails
	stockDetails.unVestedAmount = strconv.FormatFloat(float64(unvestedAmount), 'f', 5, 32)
	stockDetails.stocksBought	= stocksArray
	StockDetails[TradeID]=stockDetails
	s:= "Trade Id : " + strconv.Itoa(TradeID) +"\n" + "Stocks :  " +"'" + stocks + "'" + "\n" + "Unvested Amount : " + strconv.FormatFloat(float64(unvestedAmount), 'f', 5, 32) + "\n"
	return s
}

func callYahooServiceAPICheckPortfolio(tradeIdRequest *TradeIdRequest) string{
	var stocks string
	i,_:= strconv.Atoi(tradeIdRequest.TradeId)
	if stockBought, ok := StockDetails[i];ok{
		unVestedAmount:= stockBought.unVestedAmount
		stocksBought:= stockBought.stocksBought
		for _,stockDetails:=range stocksBought{
			stocks+=stockDetails.StockName
			stocks+=","
		}
		v:= new(YahooResponse)
		v=callYahooServiceAPI(stocks)
		responseString:= calculateLossGain(v,stocksBought)
		responseString+="\n"+"Unvested amount: "+unVestedAmount
		return responseString
	}
	return "TradeId was not found. Sorry!"
}

func calculateLossGain(response *YahooResponse,stockBought []StockBought) string{
respString := "Stocks : "
var totalCurrentStockAmount float32
for _,stocksResp:=range response.List.Resources{
	for _,stockDetails:=range stockBought{
			if stocksResp.Resource.Fields.Symbol == stockDetails.StockName{
					respString+= "`"+stockDetails.StockName+":"
					stockPrice,_:=strconv.ParseFloat(stocksResp.Resource.Fields.Price,32)
					stockPrice32:=float32(stockPrice)
					currentStockAmount:=float32(stockDetails.NumberOfStocks)*stockPrice32
					totalCurrentStockAmount=totalCurrentStockAmount+currentStockAmount
					respString+=strconv.Itoa(stockDetails.NumberOfStocks)+":"
					if stockDetails.BuyingPrice<currentStockAmount{
						respString+="+"+strconv.FormatFloat(float64(currentStockAmount), 'f', 5, 32)
					}else if stockDetails.BuyingPrice>currentStockAmount{
						respString+="-"+strconv.FormatFloat(float64(currentStockAmount), 'f', 5, 32)
					}else{
						respString+=strconv.FormatFloat(float64(currentStockAmount), 'f', 5, 32)
					}
					respString+="`,"
					break
		}
	}
}
respString = strings.TrimSuffix(respString, ",") +"\n"
respString+="Current Market Value : "+strconv.FormatFloat(float64(totalCurrentStockAmount), 'f', 5, 32)
return respString
}

func callYahooServiceAPI(stocks string) *YahooResponse{
	stocks = strings.TrimSuffix(stocks, ",")
	url := "http://finance.yahoo.com/webservice/v1/symbols/"+stocks+"/quote?format=json&view=%E2%80%8C%E2%80%8Bdetail"
	resp,err:= http.Get(url)
	if err != nil {
		log.Fatal()
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	v:=&YahooResponse{}
	json.Unmarshal(body, &v)
	return v
}
