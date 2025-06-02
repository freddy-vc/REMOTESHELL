package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

// EnviarComandos lee los comandos del usuario y los envía al servidor
func EnviarComandos(conn net.Conn, username string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s> ", username)
		comando, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer el comando:", err)
			continue
		}

		comando = strings.TrimSpace(comando)
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
	}
}

// RecibirRespuestas recibe y muestra las respuestas del servidor
func RecibirRespuestas(conn net.Conn) {
	for {
		respuesta, err := leerRespuestaCompleta(conn)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("Error al recibir respuesta:", err)
			continue
		}

		if respuesta != "" {
			fmt.Print(respuesta)
			if !strings.HasSuffix(respuesta, "\n") {
				fmt.Println()
			}
		}
	}
}

// leerRespuestaCompleta lee la respuesta completa del servidor
func leerRespuestaCompleta(conn net.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	reader := bufio.NewReader(conn)
	var respuestaCompleta strings.Builder
	buffer := make([]byte, 4096)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() && respuestaCompleta.Len() > 0 {
				break
			}
			return "", err
		}

		respuestaCompleta.Write(buffer[:n])
		if n < len(buffer) {
			break
		}
	}

	return respuestaCompleta.String(), nil
}

// ExecuteRemoteCommand maneja la ejecución de comandos remotos
func ExecuteRemoteCommand(conn net.Conn, username string) {
	fmt.Println("[x]----------|||---------[x]")
	fmt.Println("|   REMOTESHELL - CLIENT   |")
	fmt.Printf("   Bienvenido: %s", username)
	fmt.Println("\n| Escriba 'bye' para salir |")
	fmt.Println("[x]----------|||---------[x]")
	// Goroutine para recibir respuestas del servidor
	go RecibirRespuestas(conn)

	// Enviar comandos al servidor (en el hilo principal)
	EnviarComandos(conn, username)
}
