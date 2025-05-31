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
		// Solicitar parámetros de conexión en cada intento
		ip, puerto, periodoReporte, err := SolicitarParametros()
		if err != nil {
			fmt.Printf("Error al obtener parámetros: %v\n", err)
			os.Exit(1)
		}

		conn, err := Conectar(ip, puerto, periodoReporte)
		if err != nil {
			fmt.Printf("Error al conectar o autenticar con el servidor: %v\n", err)
			intentos++
			if intentos < intentosMax {
				fmt.Printf("Intento %d de %d. Intente nuevamente.\n", intentos+1, intentosMax)
			}
			continue // vuelve a pedir los parámetros de conexión
		}

		// Iniciar el envío periódico de reportes en una goroutine
		go StartReport(conn, periodoReporte)
		StartCommandShell(conn)
		conectado = true
		break // sale del bucle principal si la conexión y autenticación fueron exitosas
	}

	if !conectado && intentos >= intentosMax {
		fmt.Println("Se alcanzó el número máximo de intentos fallidos de conexión.")
		os.Exit(1)
	}
}

func SolicitarParametros() (string, string, int, error) {
	reader := bufio.NewReader(os.Stdin)

	// Solicitar IP
	fmt.Print("Ingrese la IP del servidor: ")
	ip, err := reader.ReadString('\n')
	if err != nil {
		return "", "", 0, err
	}
	ip = strings.TrimSpace(ip)

	// Solicitar Puerto
	fmt.Print("Ingrese el puerto del servidor: ")
	puerto, err := reader.ReadString('\n')
	if err != nil {
		return "", "", 0, err
	}
	puerto = strings.TrimSpace(puerto)

	// Solicitar Periodo de Reporte
	fmt.Print("Ingrese el periodo de reporte en segundos: ")
	periodoStr, err := reader.ReadString('\n')
	if err != nil {
		return "", "", 0, err
	}
	periodoStr = strings.TrimSpace(periodoStr)

	// Convertir periodo a entero
	periodo, err := strconv.Atoi(periodoStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("el periodo debe ser un número válido")
	}

	// Validar parámetros
	if ip == "" || puerto == "" || periodo <= 0 {
		return "", "", 0, fmt.Errorf("todos los parámetros son obligatorios y el periodo debe ser mayor a 0")
	}

	return ip, puerto, periodo, nil
}
