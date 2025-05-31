package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func getSystemInfo() (cpuUsage float64, memUsage float64, memFree float64, memTotal float64, diskUsage float64, diskFree float64, diskTotal float64, procCount int, err error) {
	// CPU Usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("error obteniendo CPU: %v", err)
	}
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// Memory
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("error obteniendo memoria: %v", err)
	}
	memUsage = vmStat.UsedPercent
	memFree = float64(vmStat.Available) / 1024 / 1024 // MB
	memTotal = float64(vmStat.Total) / 1024 / 1024    // MB

	// Disk
	diskStat, err := disk.Usage("C:")
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("error obteniendo disco: %v", err)
	}
	diskUsage = diskStat.UsedPercent
	diskFree = float64(diskStat.Free) / 1024 / 1024 / 1024   // GB
	diskTotal = float64(diskStat.Total) / 1024 / 1024 / 1024 // GB

	// Process Count
	processes, err := process.Processes()
	if err != nil {
		return cpuUsage, memUsage, memFree, memTotal, diskUsage, diskFree, diskTotal, 0, fmt.Errorf("error obteniendo procesos: %v", err)
	}
	procCount = len(processes)

	return cpuUsage, memUsage, memFree, memTotal, diskUsage, diskFree, diskTotal, procCount, nil
}

// StartReport envía periódicamente un reporte de consumo de recursos al servidor.
func StartReport(conn net.Conn, periodo int) {
	for {
		// Obtener toda la información del sistema
		cpuUsage, memUsage, memFree, memTotal, diskUsage, diskFree, diskTotal, procCount, err := getSystemInfo()
		if err != nil {
			fmt.Printf("Error obteniendo información del sistema: %v\n", err)
		}

		// Crear el reporte
		reporte := fmt.Sprintf("__REPORTE__: [WINDOWS] Recursos del Sistema:\n"+
			"- CPU: %.2f%%\n"+
			"- Memoria: %.2f%% (%.2f MB libre de %.2f MB)\n"+
			"- Disco C: %.2f%% (%.2f GB libre de %.2f GB)\n"+
			"- Procesos Activos: %d\n"+
			"- Hora: %s\n",
			cpuUsage,
			memUsage, memFree, memTotal,
			diskUsage, diskFree, diskTotal,
			procCount,
			time.Now().Format("2006-01-02 15:04:05"))

		// Adquirir el mutex antes de enviar el reporte
		ResponseMutex.Lock()

		// Enviar el reporte
		_, err = conn.Write([]byte(reporte))
		if err != nil {
			ResponseMutex.Unlock()
			fmt.Println("Error al enviar el reporte:", err)
			return
		}

		// Leer la respuesta del reporte
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

		// Leer la respuesta de sincronización
		_, err = reader.ReadString('\n')
		if err != nil {
			ResponseMutex.Unlock()
			fmt.Println("Error al leer respuesta de sincronización:", err)
			return
		}

		// Liberar el mutex
		ResponseMutex.Unlock()

		// Esperar el periodo especificado
		time.Sleep(time.Duration(periodo) * time.Second)
	}
}
