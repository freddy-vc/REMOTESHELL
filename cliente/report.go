package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

// StartReport recibe y presenta periódicamente el reporte de consumo de recursos del servidor.
func StartReport(conn net.Conn, username string) {
	reader := bufio.NewReader(conn)

	for {
		var reporte strings.Builder

		// Leer línea por línea hasta encontrar una línea vacía
		for {
			linea, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
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
	}
}
