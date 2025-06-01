package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func SolicitarCredenciales(socket net.Conn) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var usuario string

	for {
		// Leer el número de intentos restantes o mensaje inicial del servidor
		respuesta := make([]byte, 1024)
		n, err := socket.Read(respuesta)
		if err != nil {
			return "", fmt.Errorf("error al leer respuesta del servidor: %v", err)
		}
		fmt.Printf("%s", string(respuesta[:n]))

		// Verificar si se agotaron los intentos
		respuestaStr := strings.TrimSpace(string(respuesta[:n]))
		if respuestaStr == "MAX_ATTEMPTS" {
			return "", fmt.Errorf("se agotaron los intentos de autenticación")
		}

		// Solicitar usuario
		fmt.Print("Ingrese su nombre de usuario: ")
		usuario, err = reader.ReadString('\n')
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

		// Leer respuesta del servidor sobre el usuario
		respuesta = make([]byte, 1024)
		n, err = socket.Read(respuesta)
		if err != nil {
			return "", fmt.Errorf("error al leer respuesta del servidor: %v", err)
		}
		respuestaStr = strings.TrimSpace(string(respuesta[:n]))

		// Manejar respuestas relacionadas con el usuario
		switch respuestaStr {
		case "USER_NOT_FOUND":
			fmt.Println("Usuario no encontrado. Intente nuevamente.")
			continue
		case "USER_NOT_ALLOWED":
			fmt.Println("Usuario no permitido. Intente con otro usuario.")
			continue
		case "MAX_ATTEMPTS":
			return "", fmt.Errorf("se agotaron los intentos de autenticación")
		}

		// Si llegamos aquí, el usuario es válido, solicitar contraseña
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

		// Leer respuesta final del servidor
		respuesta = make([]byte, 1024)
		n, err = socket.Read(respuesta)
		if err != nil {
			return "", fmt.Errorf("error al leer respuesta de autenticación: %v", err)
		}
		respuestaStr = strings.TrimSpace(string(respuesta[:n]))

		// Manejar respuesta final
		switch respuestaStr {
		case "AUTH_OK":
			fmt.Println("Autenticación exitosa")
			return usuario, nil
		case "PASSWORD_ERROR":
			fmt.Println("Contraseña incorrecta. Intente nuevamente.")
			break // Volver al inicio del bucle para solicitar usuario
		case "MAX_ATTEMPTS":
			return "", fmt.Errorf("se agotaron los intentos de autenticación")
		default:
			return "", fmt.Errorf("respuesta no reconocida del servidor: %s", respuestaStr)
		}
	}
}
