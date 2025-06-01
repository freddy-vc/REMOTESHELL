package main

import (
	"fmt"
	"os"
	"strconv"
)

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

	// Intentar conectar con el servidor
	conn, username, err := Conectar(serverIP, serverPort, periodoReporte)
	if err != nil {
		fmt.Printf("Error al conectar o autenticar con el servidor: %v\n", err)
		os.Exit(1)
	} else {
		// Iniciar el envío periódico de reportes en una goroutine
		go StartReport(conn, username, periodoReporte)

		// Iniciar la shell de comandos
		StartCommandShell(conn, username)
	}
}
