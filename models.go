package main

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
    Password string `json:"password"`
	Email    string `json:"email"`
}

type UserDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func NewUserDto(uname string, pass string, email string) *UserDto {
    return &UserDto{
        Username: uname,
        Password: pass,
        Email:    email,
    }
}

func NewUser(id int, uname string, pass string, email string) *User {
    return &User{
        ID: id,
        Username: uname,
        Password: pass,
        Email: email,
    }
}
