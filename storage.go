package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Storage interface {
	CreateUser(*UserDto) error
	DeleteUser(int64) error
	UpdateUser(int64, *UserDto) (*User, error)
	GetUsers() ([]*User, error)
	GetUserByID(int64) (*User, error)
    GetUserByEmail(string) (*User, error)

	CreateStation(*StationDto) error
	DeleteStation(int64) error
	UpdateStation(int64, *StationDto) error
	GetStations() ([]*Station, error)
	GetStationByID(int64) (*Station, error)

    GetHistoryPrices(int64, string) (map[time.Time]float64, error)
    GetPricesByLocation(Location) ([]*Station, error)
}

type RAMStorage struct {
	users    []*User
	stations []*Station
	mu       sync.Mutex
}

func NewRAMStorage() *RAMStorage {
	return &RAMStorage{
		users:    make([]*User, 0),
		stations: make([]*Station, 0),
	}
}

func generateId() int64 {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return int64(r.Uint64())
}

func (s *RAMStorage) GetPricesByLocation(loc Location) ([]*Station, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    stations := make([]*Station, 0)
    for _, st := range s.stations {
        if st.Location == loc {
            stations = append(stations, st)
        }
    }
    
    return stations, nil
}

func (s *RAMStorage) GetHistoryPrices(id int64, gasType string) (map[time.Time]float64, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    station, err := s.GetStationByID(id)
    if err != nil {
        return nil, err
    }
    
    if !ValidGasType(gasType) {
        return nil, fmt.Errorf("Invalid gas type")
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
        return nil, fmt.Errorf("Gas type not supported")
    }

    histPrices := make(map[time.Time]float64, 0)
    for _, gp := range station.PricesHistory {
        histPrices[gp.Time] = gp.Prices[gt]
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
    priceModifier := NewMCPriceGen(5*time.Second, priceSource)
    priceReceiver := NewStationPriceReceiver(station)

    priceChan := make(chan GasPrices)
    go priceModifier.SendPrice(priceChan)
    go priceReceiver.ReceivePrice(priceChan)
    
	s.stations = append(s.stations, station)
	return nil
}

func (s *RAMStorage) DeleteStation(id int64) error {
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

func (s *RAMStorage) UpdateStation(id int64, station *StationDto) error {
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

func (s *RAMStorage) GetStationByID(id int64) (*Station, error) {
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

func (s *RAMStorage) DeleteUser(id int64) error {
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

func (s *RAMStorage) UpdateUser(id int64, user *UserDto) (*User, error) {
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

func (s *RAMStorage) GetUserByID(id int64) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}

	return nil, fmt.Errorf("User with id %d not found", id)
}

func (s * RAMStorage) GetUserByEmail(email string) (*User, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    for _, u := range s.users {
        if u.Email == email {
            return u, nil
        }
    }

    return nil, fmt.Errorf("User with email %s not found", email)
}
