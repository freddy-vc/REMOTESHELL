package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Función para obtener la IP local del cliente
func obtenerIPLocal() (string, error) {
	// Ejecutar el comando ipconfig
	cmd := exec.Command("ipconfig")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error al ejecutar ipconfig: %v", err)
	}

	// Convertir la salida a string y dividir por líneas
	lines := strings.Split(string(output), "\n")

	// Variables para controlar la búsqueda
	var ipWifi string
	encontradoWifi := false

	// Buscar la sección de WiFi y su IP
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Buscar el adaptador WiFi
		if strings.Contains(line, "Wi-Fi") {
			encontradoWifi = true
			continue
		}

		// Si estamos en la sección WiFi, buscar la IP
		if encontradoWifi {
			// Buscar específicamente la línea que contiene "Dirección IPv4"
			if strings.Contains(line, "IPv4") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					ip := strings.TrimSpace(parts[1])
					// Eliminar cualquier punto extra al final
					ip = strings.TrimRight(ip, ".")
					// Verificar que es una IP válida
					if net.ParseIP(ip) != nil {
						ipWifi = ip
					}
				}
			}

			// Solo terminar la sección WiFi cuando encontremos el siguiente adaptador
			// o cuando la IP ya fue encontrada
			if (strings.Contains(line, "Adaptador") && !strings.Contains(line, "Wi-Fi")) ||
				(ipWifi != "") {
				break
			}
		}
	}

	if ipWifi != "" {
		return ipWifi, nil
	}

	return "", fmt.Errorf("no se encontró la dirección IPv4 del adaptador WiFi")
}

func autenticarConServidor(socket net.Conn) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Leer el número de intentos restantes
	respuesta := make([]byte, 1024)
	n, err := socket.Read(respuesta)
	if err != nil {
		return "", fmt.Errorf("error al leer respuesta inicial: %v", err)
	}
	fmt.Printf("Intentos restantes: %s", string(respuesta[:n]))

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

func Conectar(ip string, puerto string, periodoReporte int) (net.Conn, string, error) {
	// Conectar al servidor
	direccion := net.JoinHostPort(ip, puerto)
	conn, err := net.Dial("tcp", direccion)
	if err != nil {
		return nil, "", fmt.Errorf("error al conectar con el servidor: %v", err)
	}

	// Autenticar con el servidor
	username, err := autenticarConServidor(conn)
	if err != nil {
		conn.Close()
		return nil, "", err
	}

	return conn, username, nil
}
