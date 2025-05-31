package main

import (
	"bufio"
	"fmt"
	"net"
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

func validarParametros() (string, string, int, error) {
	if len(os.Args) != 4 {
		return "", "", 0, fmt.Errorf("Uso: %s <DireccionIP> <Puerto> <TiempoReporte>\nEjemplo: %s 192.168.1.100 1625 5", os.Args[0], os.Args[0])
	}

	ip := os.Args[1]
	puerto := os.Args[2]

	// Validar que la IP tenga un formato válido
	if net.ParseIP(ip) == nil {
		return "", "", 0, fmt.Errorf("La dirección IP '%s' no es válida", ip)
	}

	// Validar que el puerto sea un número válido
	puertoint, err := strconv.Atoi(puerto)
	if err != nil || puertoint < 1 || puertoint > 65535 {
		return "", "", 0, fmt.Errorf("El puerto debe ser un número entre 1 y 65535")
	}

	// Validar y convertir el tiempo de reporte
	periodo, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return "", "", 0, fmt.Errorf("El tiempo de reporte debe ser un número válido")
	}
	if periodo <= 0 {
		return "", "", 0, fmt.Errorf("El tiempo de reporte debe ser mayor a 0")
	}

	return ip, puerto, periodo, nil
}

func main() {
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

	// Obtener y validar parámetros de línea de comandos
	ip, puerto, periodoReporte, err := validarParametros()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	intentos := 0
	conectado := false
	for intentos = 0; intentos < intentosMax; {
		conn, username, err := Conectar(ip, puerto, periodoReporte)
		if err != nil {
			fmt.Printf("Error al conectar o autenticar con el servidor: %v\n", err)
			intentos++
			if intentos < intentosMax {
				fmt.Printf("Intento %d de %d. Intente nuevamente.\n", intentos+1, intentosMax)
			}
			continue
		}

		// Iniciar el envío periódico de reportes en una goroutine
		go StartReport(conn, periodoReporte)
		StartCommandShell(conn, username)
		conectado = true
		break // sale del bucle principal si la conexión y autenticación fueron exitosas
	}

	if !conectado && intentos >= intentosMax {
		fmt.Println("Se alcanzó el número máximo de intentos fallidos de conexión.")
		os.Exit(1)
	}
}
