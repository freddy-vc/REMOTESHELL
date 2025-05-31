package main

import (
	"bufio"
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

		// Adquirir el mutex antes de enviar el reporte y la sincronización
		ResponseMutex.Lock()

		// Enviar el reporte
		_, err := conn.Write([]byte(reporte))
		if err != nil {
			ResponseMutex.Unlock()
			fmt.Println("Error al enviar el reporte:", err)
			return
		}

		// Leer la respuesta del reporte sin mostrarla
		reader := bufio.NewReader(conn)
		_, err = reader.ReadString('\n')
		if err != nil {
			ResponseMutex.Unlock()
			fmt.Println("Error al leer respuesta del reporte:", err)
			return
		}

		// Enviar comando de sincronización
		_, err = conn.Write([]byte("__SYNC__\n"))
		if err != nil {
			ResponseMutex.Unlock()
			fmt.Println("Error al enviar comando de sincronización:", err)
			return
		}

		// Leer la respuesta de sincronización sin mostrarla
		_, err = reader.ReadString('\n')
		if err != nil {
			ResponseMutex.Unlock()
			fmt.Println("Error al leer respuesta de sincronización:", err)
			return
		}

		// Liberar el mutex después de completar la sincronización
		ResponseMutex.Unlock()

		time.Sleep(time.Duration(periodo) * time.Second)
	}
}
