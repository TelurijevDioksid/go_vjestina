package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
    "strconv"
)

type APIServer struct {
	port string
    storage Storage
}

type APIError struct {
	Error string `json:"error"`
}

type apiFuncDef func(http.ResponseWriter, *http.Request) error

func jsonWriter(w http.ResponseWriter, status int, data interface{}) error {
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

func wrapJWTAuth(hFunc http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("JWT Auth")
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

	router.HandleFunc("/user", wrapJWTAuth(wrapApiHandleFunc(s.handleUser)))
	router.HandleFunc("GET /user/{id}", wrapJWTAuth(wrapApiHandleFunc(s.handleGetUserById)))
    router.HandleFunc("DELETE /user/{id}", wrapJWTAuth(wrapApiHandleFunc(s.handleDeleteUser)))

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

func (s *APIServer) handleUser(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
        case http.MethodGet:
            return s.handleGetUsers(w, r)
        case http.MethodPost:
            return s.handleCreateUser(w, r)
        case http.MethodPut:
            return s.handleUpdateUser(w, r)
        default:
            return fmt.Errorf("Method %s not allowed", r.Method)
	}
}

func (s *APIServer) handleGetUsers(w http.ResponseWriter, r *http.Request) error {
    allusers, err := s.storage.GetUsers()
    if err != nil {
        return fmt.Errorf("Failed to get users")
    }
    return jsonWriter(w, http.StatusOK, allusers)
}

func (s *APIServer) handleGetUserById(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
    intId, err := strconv.Atoi(id)
    if err != nil {
        return fmt.Errorf("Invalid id %s", id)
    }

    user, err := s.storage.GetUserByID(intId)
    if err != nil {
        return err
    }

	return jsonWriter(w, http.StatusOK, user)
}

func (s *APIServer) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
    userDto := &UserDto{}
    if err := json.NewDecoder(r.Body).Decode(userDto); err != nil {
        return err
    }

    user, err := s.storage.CreateUser(userDto) 
    if err != nil {
        return err
    }

    return jsonWriter(w, http.StatusCreated, user)
}

func (s *APIServer) handleUpdateUser(w http.ResponseWriter, r *http.Request) error {
    return nil
}

func (s *APIServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) error {
    id := r.PathValue("id")
    intId, err := strconv.Atoi(id)
    if err != nil {
        return fmt.Errorf("Invalid id %s", id)
    }

    err = s.storage.DeleteUser(intId)
    if err != nil {
        return err
    }

    return jsonWriter(w, http.StatusOK, "deleted user with id " + strconv.Itoa(intId))
}
