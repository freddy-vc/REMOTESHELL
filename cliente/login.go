package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func SolicitarCredenciales(socket net.Conn) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Leer el número de intentos restantes o mensaje inicial del servidor
	respuesta := make([]byte, 1024)
	n, err := socket.Read(respuesta)
	if err != nil {
		return "", fmt.Errorf("error al leer respuesta inicial: %v", err)
	}
	fmt.Printf("%s", string(respuesta[:n]))

	// Solicitar usuario
	fmt.Print("Ingrese su nombre de usuario: ")
	usuario, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer usuario: %v", err)
	}
	usuario = strings.TrimSpace(usuario)
	usuarioConSalto := usuario + "\n"

	// Enviar usuario al servidor
	_, err = socket.Write([]byte(usuarioConSalto))
	if err != nil {
		return "", fmt.Errorf("error al enviar usuario: %v", err)
	}

	// Solicitar contraseña
	fmt.Print("Ingrese su contraseña: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer contraseña: %v", err)
	}
	password = strings.TrimSpace(password) + "\n"

	// Enviar contraseña al servidor
	_, err = socket.Write([]byte(password))
	if err != nil {
		return "", fmt.Errorf("error al enviar contraseña: %v", err)
	}

	// Configurar un timeout para la respuesta
	socket.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer socket.SetReadDeadline(time.Time{}) // Restaurar el timeout por defecto

	// Leer respuesta del servidor
	respuesta = make([]byte, 1024)
	n, err = socket.Read(respuesta)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return "", fmt.Errorf("timeout esperando respuesta del servidor")
		}
		return "", fmt.Errorf("error al leer respuesta de autenticación: %v", err)
	}

	respuestaStr := strings.TrimSpace(string(respuesta[:n]))

	// Manejar diferentes tipos de respuestas
	switch respuestaStr {
	case "AUTH_OK":
		fmt.Println("Autenticación exitosa")
		return usuario, nil
	case "AUTH_ERROR":
		return "", fmt.Errorf("usuario o contraseña incorrectos")
	case "IP_ERROR":
		return "", fmt.Errorf("IP no autorizada")
	default:
		return "", fmt.Errorf("respuesta no reconocida del servidor: %s", respuestaStr)
	}
}
