package main

import (
	"sync"
    "fmt"
)

type Storage interface {
	CreateUser(*UserDto) (*User, error)
	DeleteUser(int) error
	UpdateUser(int, *User) error
	GetUsers() ([]*User, error)
	GetUserByID(int) (*User, error)
}

type RAMStorage struct {
	users map[int]*User
	mu    sync.Mutex
}

func NewRAMStorage() *RAMStorage {
	return &RAMStorage{
		users: make(map[int]*User),
	}
}

func (s *RAMStorage) CreateUser(u *UserDto) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastid := len(s.users)
	user := NewUser(lastid, u.Username, u.Password, u.Email)
	s.users[lastid] = user
	return user, nil
}

func (s *RAMStorage) DeleteUser(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

    _, exists := s.users[id]
    if !exists {
        return fmt.Errorf("User with id %d not found", id)
    }
	delete(s.users, id)
	return nil
}

func (s *RAMStorage) UpdateUser(id int, u *User) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.users[id] = u
    return nil
}

func (s *RAMStorage) GetUsers() ([]*User, error) {
	allusers := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		allusers = append(allusers, user)
	}
	return allusers, nil
}

func (s *RAMStorage) GetUserByID(id int) (*User, error) {
	user := s.users[id]
    if user == nil {
        return nil, fmt.Errorf("User with id %d not found", id)
    }
	return user, nil
}
