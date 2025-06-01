package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// ResponseMutex se usa para sincronizar las respuestas del servidor
var ResponseMutex sync.Mutex

// ExecuteRemoteCommand envía un comando al servidor remoto y muestra la respuesta
func ExecuteRemoteCommand(conn net.Conn, username string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s> ", username)
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

		// Adquirir el mutex antes de enviar el comando
		ResponseMutex.Lock()

		// Enviar el comando al servidor
		_, err = conn.Write([]byte(comando + "\n"))
		if err != nil {
			ResponseMutex.Unlock()
			fmt.Println("Error al enviar el comando:", err)
			continue
		}

		// Esperar y leer la respuesta del servidor
		respuesta, err := leerRespuestaCompleta(conn)

		// Liberar el mutex después de recibir la respuesta
		ResponseMutex.Unlock()

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Tiempo de espera agotado. El comando puede haberse ejecutado correctamente.")
			} else {
				fmt.Println("Error al recibir la respuesta:", err)
			}
			continue
		}

		// Mostrar la respuesta
		if respuesta != "" {
			if strings.HasPrefix(comando, "cd ") {
				fmt.Print(respuesta)
			} else {
				// Para otros comandos, asegurar que la salida sea visible
				fmt.Print(respuesta)
				if !strings.HasSuffix(respuesta, "\n") {
					fmt.Println()
				}
			}
		}
	}
}

// leerRespuestaCompleta lee la respuesta completa del servidor
func leerRespuestaCompleta(conn net.Conn) (string, error) {
	// Establecer un timeout razonable
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetReadDeadline(time.Time{}) // Restaurar el timeout por defecto

	reader := bufio.NewReader(conn)
	var respuestaCompleta strings.Builder
	buffer := make([]byte, 4096)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Si ya tenemos datos y es un timeout, consideramos que tenemos la respuesta completa
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() && respuestaCompleta.Len() > 0 {
				break
			}
			return "", err
		}

		respuestaCompleta.Write(buffer[:n])

		// Si recibimos menos datos que el tamaño del buffer, probablemente es el final
		if n < len(buffer) {
			break
		}
	}

	return respuestaCompleta.String(), nil
}

// StartCommandShell inicia el shell de comandos remoto
func StartCommandShell(conn net.Conn, username string) {
	fmt.Println("*******************************************")
	fmt.Println("*       SHELL REMOTO - CLIENTE            *")
	fmt.Printf("*       Usuario: %-24s*\n", username)
	fmt.Println("*      Escriba 'bye' para salir           *")
	fmt.Println("*******************************************")

	ExecuteRemoteCommand(conn, username)
}
