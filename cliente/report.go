package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

// StartReport recibe y presenta periódicamente el reporte de consumo de recursos del servidor.
func StartReport(conn net.Conn, periodo int) {
	for {
		// Solicitar reporte al servidor
		_, err := conn.Write([]byte("__GET_REPORT__\n"))
		if err != nil {
			fmt.Println("Error al solicitar reporte:", err)
			return
		}

		// Leer el reporte del servidor
		reader := bufio.NewReader(conn)
		reporte, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer reporte:", err)
			return
		}

		// Presentar el reporte
		fmt.Print(reporte)

		// Esperar el periodo especificado
		time.Sleep(time.Duration(periodo) * time.Second)
	}
}
