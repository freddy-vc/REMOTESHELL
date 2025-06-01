package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// StartReport recibe y presenta periódicamente el reporte de consumo de recursos del servidor.
func StartReport(conn net.Conn, username string, periodo int) {
	for {
		// Solicitar reporte al servidor
		_, err := conn.Write([]byte("__GET_REPORT__\n"))
		if err != nil {
			fmt.Printf("Error al solicitar reporte: %v\n", err)
			return
		}

		// Leer el reporte del servidor
		reader := bufio.NewReader(conn)
		var reporte strings.Builder

		// Leer línea por línea hasta encontrar una línea vacía
		for {
			linea, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error al leer reporte: %v\n", err)
				return
			}

			// Si la línea está vacía (solo contiene \n), terminar la lectura
			if strings.TrimSpace(linea) == "" {
				break
			}

			reporte.WriteString(linea)
		}

		// Presentar el reporte
		fmt.Print("\n", reporte.String())

		// Esperar el periodo especificado
		time.Sleep(time.Duration(periodo) * time.Second)
	}
}
