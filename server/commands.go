package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

var (
	currentDir string
	cmdMutex   sync.Mutex
)

func init() {
	// Inicializar el directorio actual
	currentDir, _ = os.Getwd()
}

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
	diskStat, err := disk.Usage("/")
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

func generateSystemReport() string {
	cpuUsage, memUsage, memFree, memTotal, diskUsage, diskFree, diskTotal, procCount, err := getSystemInfo()
	if err != nil {
		return fmt.Sprintf("Error generando reporte: %v\n", err)
	}

	return fmt.Sprintf("[DEBIAN] Recursos del Sistema:\n"+
		"- CPU: %.2f%%\n"+
		"- Memoria: %.2f%% (%.2f MB libre de %.2f MB)\n"+
		"- Disco: %.2f%% (%.2f GB libre de %.2f GB)\n"+
		"- Procesos Activos: %d\n"+
		"- Hora: %s\n",
		cpuUsage,
		memUsage, memFree, memTotal,
		diskUsage, diskFree, diskTotal,
		procCount,
		time.Now().Format("2006-01-02 15:04:05"))
}

func ExecuteCommand(comando string) string {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "\n" // Retorna solo un salto de línea para comandos vacíos
	}

	// Manejar solicitud de reporte
	if comando == "__GET_REPORT__" {
		return generateSystemReport()
	}

	fmt.Printf("Ejecutando comando UNIX: %s\n", comando)

	// Manejar el comando cd de forma especial
	if strings.HasPrefix(comando, "cd ") {
		// Extraer el directorio
		dir := strings.TrimPrefix(comando, "cd ")
		dir = strings.TrimSpace(dir)

		// Cambiar el directorio
		err := os.Chdir(dir)
		if err != nil {
			return fmt.Sprintf("Error al cambiar al directorio %s: %v\n", dir, err)
		}

		// Actualizar el directorio actual
		currentDir, _ = os.Getwd()
		return fmt.Sprintf("Directorio cambiado a: %s\n", currentDir)
	}

	// Ejecutar comando usando /bin/bash para mejor compatibilidad con comandos UNIX
	cmd := exec.Command("/bin/bash", "-c", comando)

	// Establecer el directorio de trabajo
	cmd.Dir = currentDir

	// Capturar tanto la salida estándar como los errores
	output, err := cmd.CombinedOutput()
	resultado := string(output)

	// Si hay error pero hay salida, mostrar la salida
	if err != nil {
		if len(resultado) > 0 {
			return resultado
		}
		return fmt.Sprintf("Error al ejecutar comando: %v\n", err)
	}

	// Si no hay salida, enviar mensaje de confirmación
	if strings.TrimSpace(resultado) == "" {
		return fmt.Sprintf("Comando '%s' ejecutado correctamente\n", comando)
	}

	// Asegurarse de que la salida termine con un salto de línea
	if !strings.HasSuffix(resultado, "\n") {
		resultado += "\n"
	}

	return resultado
}
