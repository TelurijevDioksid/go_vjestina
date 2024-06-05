package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type PriceSource interface {
	GetPrice() GasPrices 
}

type StationPriceSource struct {
	Station *Station
}

func NewStationPriceSource(s *Station) *StationPriceSource {
	return &StationPriceSource{
		Station: s,
	}
}

func (s *StationPriceSource) GetPrice() GasPrices {
	return s.Station.CurrentPrice
}

type PriceModifier interface {
	ModifyPrice(float64) float64
    SendPrice(ch GasPrices)
}

type MCPriceGen struct {
	interval          time.Duration
	initalPriceSource *StationPriceSource
}

func NewMCPriceGen(interval time.Duration, ps *StationPriceSource) *MCPriceGen {
	return &MCPriceGen{
		interval:          interval,
		initalPriceSource: ps,
	}
}

func (mc *MCPriceGen) ModifyPrice(num float64) float64 {
    src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	mu := 0.05
	sigma := 0.2
	T := 1.0
	dt := 0.01

	numSteps := int(T / dt)
	price := num
	for i := 0; i < numSteps; i++ {
		dW := math.Sqrt(dt) * rnd.NormFloat64()
		price *= math.Exp((mu - 0.5 * sigma * sigma) * dt + sigma * dW)
	}
    
    fmt.Println("Modify price: (", price, " old price ", num, ")")
    return price
}

func (mc *MCPriceGen) SendPrice(ch chan GasPrices) {
	for {
        fmt.Println("GO Generating new price")
		prevPrice := mc.initalPriceSource.GetPrice()
        newPrice := GasPrices{
            Prices: make(map[GasType]float64),
            Time: time.Now(),
        }
        for k, v := range prevPrice.Prices {
            fmt.Println("k: ", k, " v: ", v)
            newPrice.Prices[k] = mc.ModifyPrice(v)
        }
        ch <- newPrice
		time.Sleep(mc.interval)
	}
}

type PriceReceiver interface {
    ReceivePrice(ch chan map[GasType]float64)
}

type StationPriceReceiver struct {
    Station *Station
}

func NewStationPriceReceiver(s *Station) *StationPriceReceiver {
    return &StationPriceReceiver{
        Station: s,
    }
}

func (s *StationPriceReceiver) ReceivePrice(ch chan GasPrices) {
    for {
        fmt.Println("GO Receiving new price")
        newPrice := <-ch
        s.Station.PricesHistory = append(s.Station.PricesHistory, s.Station.CurrentPrice)
        s.Station.CurrentPrice = newPrice
    }
}
