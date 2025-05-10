package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	fmt.Println("************************************")
	fmt.Println("*    SERVIDOR proy. oper 2025      *")
	fmt.Println("************************************")

<<<<<<< HEAD
	socketInicial, _ := net.Listen("tcp", "192.168.1.54:1625")
	fmt.Println("Soocket  creado - OK")
=======
	// Crear el socket y escuchar en el puerto 1625
	socketInicial, err := net.Listen("tcp", ":1625")
	if err != nil {
		fmt.Println("Error al crear el socket:", err)
		return
	}
	fmt.Println("Socket creado - OK")
>>>>>>> 2fec122068e727ece7b46f2affeb19e04a155ef1
	fmt.Println("Esperando Conexiones...")

	// Aceptar una sola conexión
	socket, err := socketInicial.Accept()
	if err != nil {
		fmt.Println("Error al aceptar conexión:", err)
		return
	}
	fmt.Println("Cliente Conectado:", socket.RemoteAddr())

	// Mantener la conexión activa leyendo mensajes del cliente
	reader := bufio.NewReader(socket)
	for {
		mensaje, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("El cliente se desconectó.")
			break
		}
		fmt.Print("Mensaje recibido: ", mensaje)

		// Opcional: responder al cliente
		socket.Write([]byte("Mensaje recibido\n"))
	}
}
