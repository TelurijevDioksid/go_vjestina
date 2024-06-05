package main

import "os"

func main() {
    os.Setenv("JWT_SECRET", "GOGOGOGO")
    os.Setenv("ADMIN_UNAME", "admin")
    os.Setenv("ADMIN_PASS", "admin")
    os.Setenv("ADMIN_EMAIL", "admin@email.go")

    ramstore := NewRAMStorage()
    server := NewAPIServer(":8080", ramstore)
    server.Start()
}
