package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
    "strings"
)

type APIServer struct {
	port    string
	storage Storage
}

type APIError struct {
	Error string `json:"error"`
}

type apiFuncDef func(http.ResponseWriter, *http.Request) error

func jsonWriter(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "*")
    w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func wrapApiHandleFunc(f apiFuncDef) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			jsonWriter(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func getIdFromPath(r *http.Request) (uint64, error) {
    id_param := r.PathValue("id")
    id, err := strconv.ParseUint(id_param, 10, 64)
    if err != nil {
        return 0, fmt.Errorf("Failed to parse id")
    }
    return id, nil
}

func wrapAuth(hFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        auth := r.Header.Get("Authorization")
        parts := strings.Split(auth, " ")

        if len(parts) != 2 {
            jsonWriter(w, http.StatusUnauthorized, APIError{Error: "Unauthorized"})
            return
        }

        schema := parts[0]
        token := parts[1]
        
        if schema != "Bearer" {
            jsonWriter(w, http.StatusUnauthorized, APIError{Error: "Unauthorized"})
            return
        }

        if !ValidateJwt(token) {
            jsonWriter(w, http.StatusUnauthorized, APIError{Error: "Unauthorized"})
            return
        }

        hFunc(w, r)
	}
}

func NewAPIServer(port string, storage Storage) *APIServer {
	return &APIServer{
		port:    port,
		storage: storage,
	}
}

func (s *APIServer) Start() {
	cert, err := tls.LoadX509KeyPair("tlscert/MyCertificate.crt", "tlscert/MyKey.key")
	if err != nil {
		log.Fatalln("Failed to load key pair: ", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	router := http.NewServeMux()
    
    router.HandleFunc("POST /login", wrapApiHandleFunc(s.handleLogin))
	router.HandleFunc("GET /user", wrapAuth(wrapApiHandleFunc(s.handleGetUsers)))
	router.HandleFunc("GET /user/{id}", wrapAuth(wrapApiHandleFunc(s.handleGetUserById)))
	router.HandleFunc("POST /user", wrapAuth(wrapApiHandleFunc(s.handleCreateUser)))
	router.HandleFunc("PUT /user", wrapAuth(wrapApiHandleFunc(s.handleUpdateUser)))
	router.HandleFunc("DELETE /user/{id}", wrapAuth(wrapApiHandleFunc(s.handleDeleteUser)))

    router.HandleFunc("GET /station", wrapAuth(wrapApiHandleFunc(s.handleGetStations)))
    router.HandleFunc("GET /station/{id}", wrapAuth(wrapApiHandleFunc(s.handleGetStationById)))
    router.HandleFunc("POST /station", wrapAuth(wrapApiHandleFunc(s.handleCreateStation)))
    router.HandleFunc("PUT /station", wrapAuth(wrapApiHandleFunc(s.handleUpdateStation)))
    router.HandleFunc("DELETE /station/{id}", wrapAuth(wrapApiHandleFunc(s.handleDeleteStation)))

    router.HandleFunc("GET /prices/history/{id}/{gasType}", wrapAuth(wrapApiHandleFunc(s.handleGetHistoryPrices)))
    router.HandleFunc("POST /prices/location", wrapAuth(wrapApiHandleFunc(s.handleGetPricesByLocation)))

	server := &http.Server{
		Addr:      s.port,
		Handler:   router,
		TLSConfig: config,
	}

	log.Println("Starting server on", s.port)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalln("Failed to start server, err: ", err)
	}
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
    loginDto := new(LoginDto)
    if err := json.NewDecoder(r.Body).Decode(loginDto); err != nil {
        return err
    }

    user, err := s.storage.GetUserByEmail(loginDto.Email)
    if err != nil {
        return err
    }

    if !ValidatePassword(user.CryptPassword, loginDto.Password) {
        return fmt.Errorf("Incorrect email or password")
    }

    token := GenerateJwtToken(user.Email)
    tokenDto := NewTokenDto(token)

    w.Header().Set("Authorization", token)
    return jsonWriter(w, http.StatusOK, tokenDto)
}

func (s *APIServer) handleGetUsers(w http.ResponseWriter, r *http.Request) error {
	users, err := s.storage.GetUsers()
	if err != nil {
		return fmt.Errorf("Failed to get users")
	}

	return jsonWriter(w, http.StatusOK, users)
}

func (s *APIServer) handleGetUserById(w http.ResponseWriter, r *http.Request) error {
	id, err := getIdFromPath(r)
    if err != nil {
        return err
    }

	user, err := s.storage.GetUserByID(id)
	if err != nil {
		return err
	}

	return jsonWriter(w, http.StatusOK, user)
}

func (s *APIServer) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
	userDto := new(UserDto)
	if err := json.NewDecoder(r.Body).Decode(userDto); err != nil {
		return err
	}

	if err := s.storage.CreateUser(userDto); err != nil {
		return err
	}

	return jsonWriter(w, http.StatusCreated, "User created")
}

func (s *APIServer) handleUpdateUser(w http.ResponseWriter, r *http.Request) error {
	id, err := getIdFromPath(r)
    if err != nil {
        return err
    }
    
    userDto := new(UserDto)
    if err := json.NewDecoder(r.Body).Decode(userDto); err != nil {
        return err
    }

    user, err := s.storage.UpdateUser(id, userDto)
    if err != nil {
        return err
    }

    return jsonWriter(w, http.StatusOK, user)
}

func (s *APIServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) error {
    id, err := getIdFromPath(r)
    if err != nil {
        return err
    }

	if err := s.storage.DeleteUser(id); err != nil {
		return err
	}

	return jsonWriter(w, http.StatusOK, fmt.Sprintf("User with id %d deleted", id))
}


func (s *APIServer) handleGetStations(w http.ResponseWriter, r *http.Request) error {
    stations, err := s.storage.GetStations()
    if err != nil {
        return fmt.Errorf("Failed to get stations")
    }

    return jsonWriter(w, http.StatusOK, stations)
}

func (s *APIServer) handleGetStationById(w http.ResponseWriter, r *http.Request) error {
    id, err := getIdFromPath(r)
    if err != nil {
        return err
    }

    station, err := s.storage.GetStationByID(id)
    if err != nil {
        return err
    }

    return jsonWriter(w, http.StatusOK, station)
}

func (s *APIServer) handleCreateStation(w http.ResponseWriter, r *http.Request) error {
    stationDto := new(StationDto)
    if err := json.NewDecoder(r.Body).Decode(stationDto); err != nil {
        return err
    }

    for _, fuel := range stationDto.SupportedFuel {
        if !ValidGasType(string(fuel)) {
            return fmt.Errorf("Invalid fuel type")
        }
    }

    for k, v := range stationDto.CurrentPrice {
        if v < 0 {
            return fmt.Errorf("Invalid price, has negative value")
        }
        foundFuel := false
        for _, fuel := range stationDto.SupportedFuel {
            if k == fuel {
                foundFuel = true
                break
            }
        }
        if !foundFuel {
            return fmt.Errorf("Fuel type not specified in supported fuel types")
        }
    }

    err := s.storage.CreateStation(stationDto)
    if err != nil {
        return err
    }

    return jsonWriter(w, http.StatusCreated, "Station created")
}

func (s *APIServer) handleUpdateStation(w http.ResponseWriter, r *http.Request) error {
    id, err := getIdFromPath(r)
    if err != nil {
        return err
    }

    stationDto := new(StationDto)
    if err := json.NewDecoder(r.Body).Decode(stationDto); err != nil {
        return err
    }

    if err := s.storage.UpdateStation(id, stationDto); err != nil {
        return err
    }

    return jsonWriter(w, http.StatusOK, "Station updated")
}

func (s *APIServer) handleDeleteStation(w http.ResponseWriter, r *http.Request) error {
    id, err := getIdFromPath(r)
    if err != nil {
        return err
    }

    if err := s.storage.DeleteStation(id); err != nil {
        return err
    }

    return jsonWriter(w, http.StatusOK, fmt.Sprintf("Station with id %d deleted", id))
}

func (s *APIServer) handleGetHistoryPrices(w http.ResponseWriter, r *http.Request) error {
    id, err := getIdFromPath(r)
    if err != nil {
        return err
    }

    gasType := r.PathValue("gasType")
    if gasType == "" {
        return fmt.Errorf("Gas type is required")
    }

    prices, err := s.storage.GetHistoryPrices(id, gasType)
    if err != nil {
        return err
    }

    return jsonWriter(w, http.StatusOK, prices)
}

func (s *APIServer) handleGetPricesByLocation(w http.ResponseWriter, r *http.Request) error {
    loc := new(Location)
    if err := json.NewDecoder(r.Body).Decode(loc); err != nil {
        return err
    }
    
    prices, err := s.storage.GetPricesByLocation(loc)
    if err != nil {
        return nil
    }

    return jsonWriter(w, http.StatusOK, prices)
}

