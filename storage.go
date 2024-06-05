package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
    "os"
)

type Storage interface {
	CreateUser(*UserDto) error
	DeleteUser(uint64) error
	UpdateUser(uint64, *UserDto) (*User, error)
	GetUsers() ([]*User, error)
	GetUserByID(uint64) (*User, error)
	GetUserByEmail(string) (*User, error)

	CreateStation(*StationDto) error
	DeleteStation(uint64) error
	UpdateStation(uint64, *StationDto) error
	GetStations() ([]*Station, error)
	GetStationByID(uint64) (*Station, error)

	GetHistoryPrices(uint64, string) (*HistPriceGasTypeDto, error)
	GetPricesByLocation(*Location) ([3]*StationPriceLocDto, error)
}

type RAMStorage struct {
	users    []*User
	stations []*Station
	mu       sync.Mutex
}

func NewRAMStorage() *RAMStorage {
    id := generateId()
    uname := os.Getenv("ADMIN_UNAME")
    pass := os.Getenv("ADMIN_PASS")
    email := os.Getenv("ADMIN_EMAIL")
    admin, err := NewUser(id, uname, pass, email)
    if err != nil {
        log.Fatalf("Failed to create admin user: %v", err)
    }

    users := []*User{admin}
	return &RAMStorage{
		users:    users,
		stations: make([]*Station, 0),
	}
}

func generateId() uint64 {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return r.Uint64()
}

func (s *RAMStorage) GetPricesByLocation(loc *Location) ([3]*StationPriceLocDto, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dtoArr := [3]*StationPriceLocDto{}

	for i, st := range s.stations {
		if len(dtoArr) < 3 {
			dtoArr[i] = NewStationPriceLocDto(st.Name, st.Address, st.Location, st.CurrentPrice)
			continue
		}

		d := DistanceKm(&st.Location, loc)
		for j, ss := range dtoArr {
			if d < DistanceKm(&ss.Location, loc) {
				dtoArr[j] = NewStationPriceLocDto(st.Name, st.Address, st.Location, st.CurrentPrice)
				break
			}
		}
	}

	return dtoArr, nil
}

func (s *RAMStorage) GetHistoryPrices(id uint64, gasType string) (*HistPriceGasTypeDto, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	histPrices := &HistPriceGasTypeDto{
		HistoryPrices: make(map[time.Time]float64, 0),
	}

	station, err := s.GetStationByID(id)
	if err != nil {
		return histPrices, err
	}

	if !ValidGasType(gasType) {
		return histPrices, fmt.Errorf("Invalid gas type")
	}

	gt := GasType(gasType)
	supported := false
	for _, sf := range station.SupportedFuel {
		if sf == gt {
			supported = true
			break
		}
	}
	if !supported {
		return histPrices, fmt.Errorf("Gas type not supported")
	}

	for _, gp := range station.PricesHistory {
		histPrices.HistoryPrices[gp.Time] = gp.Prices[gt]
	}

	return histPrices, nil
}

func (s *RAMStorage) CreateStation(cst *StationDto) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateId()
	histP := make([]GasPrices, 0)

	station := NewStation(
		id,
		cst.Name,
		cst.Address,
		cst.SupportedFuel,
		cst.Location,
		cst.CurrentPrice,
		histP,
	)

	priceSource := NewStationPriceSource(station)
	priceModifier := NewMCPriceGen(10*time.Second, priceSource)
	priceReceiver := NewStationPriceReceiver(station)

	priceChan := make(chan GasPrices)
	go priceModifier.SendPrice(priceChan)
	go priceReceiver.ReceivePrice(priceChan)

	s.stations = append(s.stations, station)
	return nil
}

func (s *RAMStorage) DeleteStation(id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, st := range s.stations {
		if st.ID == id {
			s.stations = append(s.stations[:i], s.stations[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("Station with id %d not found", id)
}

func (s *RAMStorage) UpdateStation(id uint64, station *StationDto) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, st := range s.stations {
		if st.ID == id {
			st.Name = station.Name
			st.Address = station.Address
			st.SupportedFuel = station.SupportedFuel
			st.Location = station.Location
			return nil
		}
	}

	return fmt.Errorf("Station with id %d not found", id)
}

func (s *RAMStorage) GetStations() ([]*Station, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stations, nil
}

func (s *RAMStorage) GetStationByID(id uint64) (*Station, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, st := range s.stations {
		if st.ID == id {
			return st, nil
		}
	}

	return nil, fmt.Errorf("Station with id %d not found", id)
}

func (s *RAMStorage) CreateUser(u *UserDto) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateId()
	user, err := NewUser(id, u.Username, u.Password, u.Email)
	if err != nil {
		return err
	}
	s.users = append(s.users, user)
	return nil
}

func (s *RAMStorage) DeleteUser(id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, u := range s.users {
		if u.ID == id {
			s.users = append(s.users[:i], s.users[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("User with id %d not found", id)
}

func (s *RAMStorage) UpdateUser(id uint64, user *UserDto) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.ID == id {
			u.Username = user.Username
			u.CryptPassword = user.Password
			u.Email = user.Email
			return u, nil
		}
	}

	return nil, fmt.Errorf("User with id %d not found", id)
}

func (s *RAMStorage) GetUsers() ([]*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.users, nil
}

func (s *RAMStorage) GetUserByID(id uint64) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}

	return nil, fmt.Errorf("User with id %d not found", id)
}

func (s *RAMStorage) GetUserByEmail(email string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, fmt.Errorf("User with email %s not found", email)
}
