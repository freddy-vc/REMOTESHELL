package main

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func executeSystemCommand(comando string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", comando)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getSystemInfo(username string) (string, error) {
	// CPU Usage usando vmstat
	cpuCmd := "vmstat 1 2 | tail -1 | awk '{print 100 - $15 \"%\"}'"
	cpuStr, err := executeSystemCommand(cpuCmd)
	if err != nil {
		return "", fmt.Errorf("error obteniendo CPU: %v", err)
	}
	// Remover el símbolo % para la conversión
	cpuStr = strings.TrimRight(cpuStr, "%")
	cpuUsage, _ := strconv.ParseFloat(cpuStr, 64)

	// Memoria usando free
	memCmd := "free -m | grep 'Mem:'"
	memStr, err := executeSystemCommand(memCmd)
	if err != nil {
		return "", fmt.Errorf("error obteniendo memoria: %v", err)
	}
	memFields := strings.Fields(memStr)
	var memTotal, memFree float64
	var memUsage float64
	if len(memFields) >= 4 {
		memTotal, _ = strconv.ParseFloat(memFields[1], 64)
		memFree, _ = strconv.ParseFloat(memFields[3], 64)
		if memTotal > 0 {
			memUsage = 100 * (1 - memFree/memTotal)
		}
	}

	// Disco usando df
	diskCmd := "df -h / | tail -n 1"
	diskStr, err := executeSystemCommand(diskCmd)
	if err != nil {
		return "", fmt.Errorf("error obteniendo disco: %v", err)
	}
	diskFields := strings.Fields(diskStr)
	var diskTotal, diskFree float64
	var diskUsage float64
	if len(diskFields) >= 5 {
		diskTotal, _ = strconv.ParseFloat(strings.TrimRight(diskFields[1], "G"), 64)
		diskFree, _ = strconv.ParseFloat(strings.TrimRight(diskFields[3], "G"), 64)
		diskUsageStr := strings.TrimRight(diskFields[4], "%")
		diskUsage, _ = strconv.ParseFloat(diskUsageStr, 64)
	}

	// Procesos activos usando ps
	procCmd := "ps aux | wc -l"
	procStr, err := executeSystemCommand(procCmd)
	if err != nil {
		return "", fmt.Errorf("error obteniendo procesos: %v", err)
	}
	procCount, _ := strconv.ParseInt(procStr, 10, 64)
	procCount-- // Restamos 1 por el header de ps

	// Formatear el reporte con los datos en tiempo real
	report := fmt.Sprintf("%s> [DEBIAN] Recursos del Sistema:\n"+
		"- CPU: %.2f%%\n"+
		"- Memoria: %.2f%% (%.2f MB libre de %.2f MB)\n"+
		"- Disco: %.2f%% (%.2f GB libre de %.2f GB)\n"+
		"- Procesos Activos: %d\n"+
		"- Hora: %s\n\n",
		username,
		cpuUsage,
		memUsage, memFree, memTotal,
		diskUsage, diskFree, diskTotal,
		procCount,
		time.Now().Format("2006-01-02 15:04:05"))

	return report, nil
}

func generateSystemReport(username string) string {
	report, err := getSystemInfo(username)
	if err != nil {
		return fmt.Sprintf("Error generando reporte: %v\n\n", err)
	}

	fmt.Print("Enviando reporte al cliente:\n", report) // Log en servidor
	return report
}

func iniciarSistemaReportes(ctx context.Context) {
	// Esta función se ejecutará como una goroutine desde iniciarServidor
	fmt.Println("Sistema de reportes iniciado")
}
