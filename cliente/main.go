package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func LeerConfigIntentos1(ruta string) (int, string, error) {
	file, err := os.Open(ruta)
	if err != nil {
		return 3, "", nil // Valor por defecto si no se encuentra el archivo
	}
	defer file.Close()

	var intentos int = 3 // Valor por defecto
	var ipPermitida string = ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "INTENTOS_MAX=") {
			fmt.Sscanf(line, "INTENTOS_MAX=%d", &intentos)
		} else if strings.HasPrefix(line, "IP_CLIENTE=") {
			ipPermitida = strings.TrimPrefix(line, "IP_CLIENTE=")
		}
	}
	return intentos, ipPermitida, nil
}

func main() {
	// Verificar argumentos
	if len(os.Args) != 4 {
		fmt.Println("Uso: ./client <IP> <Puerto> <PeriodoReporte>")
		fmt.Println("Ejemplo: ./client 10.1.10.3 2025 5")
		os.Exit(1)
	}

	serverIP := os.Args[1]
	serverPort := os.Args[2]
	periodoReporte, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Printf("Error: el periodo de reporte debe ser un número entero: %v\n", err)
		os.Exit(1)
	}

	// Leer configuración y validar IP del cliente
	intentosMax, ipPermitida, _ := LeerConfigIntentos1("../server/config.conf")

	// Verificar si la IP local coincide con la IP permitida
	if ipPermitida != "" {
		ipLocal, err := obtenerIPLocal()
		if err != nil {
			fmt.Printf("Error al obtener IP local: %v\n", err)
		} else {
			fmt.Printf("IP local: %s, IP permitida: %s\n", ipLocal, ipPermitida)
			if ipLocal != ipPermitida {
				fmt.Printf("Error: La IP local (%s) no coincide con la IP permitida (%s)\n", ipLocal, ipPermitida)
				fmt.Println("Terminando el programa por seguridad...")
				os.Exit(1)
			}
		}
	}

	intentos := 0
	conectado := false
	for intentos = 0; intentos < intentosMax; {
		conn, username, err := Conectar(serverIP, serverPort, periodoReporte)
		if err != nil {
			fmt.Printf("Error al conectar o autenticar con el servidor: %v\n", err)
			intentos++
			if intentos < intentosMax {
				fmt.Printf("Intento %d de %d. Intente nuevamente.\n", intentos+1, intentosMax)
			}
			continue
		}

		// Iniciar el envío periódico de reportes en una goroutine
		go StartReport(conn, username, periodoReporte)
		StartCommandShell(conn, username)
		conectado = true
		break
	}

	if !conectado && intentos >= intentosMax {
		fmt.Println("Se alcanzó el número máximo de intentos fallidos de conexión.")
		os.Exit(1)
	}
}
