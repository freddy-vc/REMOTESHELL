package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	fmt.Println("************************************")
	fmt.Println("*    SERVIDOR proy. oper 2025      *")
	fmt.Println("************************************")

	socketInicial, _ := net.Listen("tcp", "192.168.1.54:1625")
	fmt.Println("Socket creado - OK")
	fmt.Println("Esperando Conexiones...")
	defer socketInicial.Close()

	for {
		socket, err := socketInicial.Accept()
		if err != nil {
			fmt.Println("Error al aceptar conexión:", err)
			continue
		}

		fmt.Printf("Cliente conectado desde: %s\n", socket.RemoteAddr())
		go manejarCliente(socket)
	}
}

func manejarCliente(socket net.Conn) {
	defer socket.Close()
	reader := bufio.NewReader(socket)

	for {
		// Leer comando del cliente
		comando, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Cliente %s desconectado\n", socket.RemoteAddr())
			return
		}

		comando = strings.TrimSpace(comando)
		fmt.Printf("Comando recibido de %s: %s\n", socket.RemoteAddr(), comando)

		// Ejecutar el comando
		respuesta := ExecuteCommand(comando)

		// Enviar respuesta al cliente
		_, err = socket.Write([]byte(respuesta))
		if err != nil {
			fmt.Printf("Error al enviar respuesta a %s: %v\n", socket.RemoteAddr(), err)
			return
		}
	}
}
