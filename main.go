package main

func main() {
    ramstore := NewRAMStorage()
    server := NewAPIServer(":8080", ramstore)
    server.Start()
}
