package main

import (
	"time"
    "math"
)

type GasType string

const (
    EarthRadius = 6371
	Diesel   = "diesel"
	Gasoline = "gasoline"
	Gas      = "gas"
)

func ValidGasType(g string) bool {
	return g == Diesel || g == Gasoline || g == Gas
}

type HistPriceGasTypeDto struct {
    HistoryPrices map[time.Time]float64 `json:"history_prices"`
}

type User struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	CryptPassword string `json:"password"`
	Email         string `json:"email"`
}

type UserDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenDto struct {
	Token string `json:"token"`
}

func NewTokenDto(token string) *TokenDto {
	return &TokenDto{
		Token: token,
	}
}

func NewUser(id int64, uname string, pass string, email string) (*User, error) {
	encryPwd, err := BcryptPassword(pass)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:            id,
		Username:      uname,
		CryptPassword: encryPwd,
		Email:         email,
	}, nil
}

func NewUserDto(uname string, pass string, email string) *UserDto {
	return &UserDto{
		Username: uname,
		Password: pass,
		Email:    email,
	}
}

type Station struct {
	ID            int64       `json:"id"`
	Name          string      `json:"name"`
	Address       string      `json:"address"`
	SupportedFuel []GasType   `json:"supported_fuel"`
	Location      Location    `json:"location"`
	CurrentPrice  GasPrices   `json:"current_price"`
	PricesHistory []GasPrices `json:"price_history"`
}

type GasPrices struct {
	Prices map[GasType]float64 `json:"prices"`
	Time   time.Time           `json:"time"`
}

type StationDto struct {
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	SupportedFuel []GasType `json:"supported_fuel"`
	Location      Location  `json:"location"`
	CurrentPrice  GasPrices `json:"current_price"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewStation(
	id int64,
	name string,
	addr string,
	suppFuel []GasType,
	loc Location,
	currP GasPrices,
	histP []GasPrices) *Station {
	return &Station{
		ID:            id,
		Name:          name,
		Address:       addr,
		SupportedFuel: suppFuel,
		Location:      loc,
		CurrentPrice:  currP,
		PricesHistory: histP,
	}
}

func NewStationDto(
	name string,
	addr string,
	suppFuel []GasType,
	location Location) *StationDto {
	return &StationDto{
		Name:          name,
		Address:       addr,
		SupportedFuel: suppFuel,
		Location:      location,
	}
}

func DistanceKm(aLoc, bLoc *Location) float64 {
    lonA := aLoc.Longitude * math.Pi / 180
    lonB := bLoc.Longitude * math.Pi / 180
    latA := aLoc.Latitude * math.Pi / 180
    latB := bLoc.Latitude * math.Pi / 180
   
    // Haversine formula 
    dlon := lonB - lonA 
    dlat := latB - latA
    a := math.Pow(math.Sin(dlat / 2), 2) +
        math.Cos(latA) * math.Cos(latB) *
        math.Pow(math.Sin(dlon / 2), 2)
           
    c := 2 * math.Asin(math.Sqrt(a));

    return c * EarthRadius;
}
