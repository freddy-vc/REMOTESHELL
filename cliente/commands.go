package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// ExecuteRemoteCommand envía un comando al servidor remoto y muestra la respuesta
func ExecuteRemoteCommand(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("shell> ")
		comando, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer el comando:", err)
			continue
		}

		// Eliminar el salto de línea al final
		comando = strings.TrimSpace(comando)

		// Verificar si es el comando de salida
		if comando == "bye" {
			fmt.Println("Cerrando sesión remota...")
			return
		}

		// Enviar el comando al servidor
		_, err = conn.Write([]byte(comando + "\n"))
		if err != nil {
			fmt.Println("Error al enviar el comando:", err)
			continue
		}

		// Recibir la respuesta del servidor
		buffer := make([]byte, 4096)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error al recibir la respuesta:", err)
			continue
		}

		// Mostrar la respuesta
		fmt.Println(string(buffer[:n]))
	}
}

// StartCommandShell inicia el shell de comandos remoto
func StartCommandShell(conn net.Conn) {
	fmt.Println("*******************************************")
	fmt.Println("*       SHELL REMOTO - CLIENTE            *")
	fmt.Println("*      Escriba 'bye' para salir           *")
	fmt.Println("*******************************************")

	ExecuteRemoteCommand(conn)
}
