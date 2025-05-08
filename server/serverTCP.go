package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("************************************")
	fmt.Println("*    SERVIDOR proy. oper 2025      *")
	fmt.Println("************************************")

	socketInicial, _ := net.Listen("tcp", "192.168.137.49:1625")
	fmt.Println("Soocket  creado - OK")
	fmt.Println("Esperando Conexiones...")
	socket, _ := socketInicial.Accept()
	fmt.Println("Cliente Conectado", socket.RemoteAddr())
}
