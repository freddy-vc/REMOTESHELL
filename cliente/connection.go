package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("********************************")
	fmt.Println("*   CLIENTE PROY. OPER 2025   *")
	fmt.Println("********************************")

	socket, _ := net.Dial("tcp", "192.168.137.1:1625")
	fmt.Println("Conectado al socket: ", socket.RemoteAddr().String())
}
