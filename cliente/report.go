package main

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

// StartReport envía periódicamente un reporte de consumo de recursos al servidor.
func StartReport(conn net.Conn, periodo int) {
	var lastNumGC uint32
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Memoria en MB
		memMB := float64(m.Alloc) / 1024.0 / 1024.0

		// Número de CPUs y goroutines
		numCPU := runtime.NumCPU()
		numGoroutine := runtime.NumGoroutine()

		// Número de GC realizados
		numGC := m.NumGC
		gcDelta := numGC - lastNumGC
		lastNumGC = numGC

		reporte := fmt.Sprintf("__REPORTE__: Recursos - Memoria: %.2f MB | CPUs: %d | Goroutines: %d | GC recientes: %d | Hora: %s\n",
			memMB, numCPU, numGoroutine, gcDelta, time.Now().Format("2006-01-02 15:04:05"))

		_, err := conn.Write([]byte(reporte))
		if err != nil {
			fmt.Println("Error al enviar el reporte:", err)
			return
		}
		time.Sleep(time.Duration(periodo) * time.Second)
	}
}