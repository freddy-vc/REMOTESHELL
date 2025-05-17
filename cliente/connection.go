package main

import (
	"fmt"
	"net"
	"os"
	"bufio"
	"strings"
)

type Config struct {
	IntentosFallidos int
}

func LeerConfigIntentos(ruta string) (int, error) {
	file, err := os.Open(ruta)
	if err != nil {
		return 3, nil // Valor por defecto si no se encuentra el archivo
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "INTENTOS_MAX=") {
			var intentos int
			fmt.Sscanf(line, "INTENTOS_MAX=%d", &intentos)
			return intentos, nil
		}
	}
	return 3, nil // Valor por defecto si no se encuentra la clave
}

func Conectar(ip string, puerto string, periodoReporte int) (net.Conn, error) {
	fmt.Println("********************************")
	fmt.Println("*   CLIENTE PROY. OPER 2025    *")
	fmt.Println("********************************")

	var conn string = ip + ":" + puerto
	socket, err := net.Dial("tcp", conn)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar al servidor: %v", err)
	}
	fmt.Println("Conectado al socket: ", socket.RemoteAddr().String())

	// Leer intentos máximos desde el archivo de configuración del servidor
	intentosMax, _ := LeerConfigIntentos("../server/config.conf")
	intentos := 0
	for intentos < intentosMax {
		usuario, password, err := SolicitarCredenciales()
		if err == nil {
			fmt.Printf("Usuario %s  con contraseña %s autenticado. Periodo de reporte: %d segundos\n", usuario, password, periodoReporte)
			return socket, nil
		}
		fmt.Println("Error de autenticación:", err)
		intentos++
		if intentos < intentosMax {
			fmt.Printf("Intento %d de %d. Intente nuevamente.\n", intentos+1, intentosMax)
		}
	}
	return nil, fmt.Errorf("se alcanzó el número máximo de intentos fallidos de autenticación")
}
