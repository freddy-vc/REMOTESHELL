package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("********************************")
	fmt.Println("*   CLIENTE PROY. OPER 2025   *")
	fmt.Println("********************************")

	// Cambia la IP si el servidor está en otra máquina
	socket, err := net.Dial("tcp", "192.168.137.1:1625")
	if err != nil {
		fmt.Println("Error al conectar con el servidor:", err)
		return
	}
	defer socket.Close()

	fmt.Println("Conectado al socket:", socket.RemoteAddr().String())

	readerConsola := bufio.NewReader(os.Stdin)
	readerServidor := bufio.NewReader(socket)

	for {
		fmt.Print("Escribe un mensaje: ")
		mensaje, _ := readerConsola.ReadString('\n')

		// Enviar al servidor
		_, err := socket.Write([]byte(mensaje))
		if err != nil {
			fmt.Println("Error al enviar mensaje:", err)
			break
		}

		// Esperar y mostrar respuesta del servidor
		respuesta, err := readerServidor.ReadString('\n')
		if err != nil {
			fmt.Println("Servidor desconectado.")
			break
		}
		fmt.Println("Servidor dice:", respuesta)
	}
}
