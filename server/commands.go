package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	currentDir string
	cmdMutex   sync.Mutex
)

func init() {
	// Inicializar el directorio actual
	currentDir, _ = os.Getwd()
}

func executeSystemCommand(comando string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", comando)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getSystemInfo() (cpuUsage float64, memUsage float64, memFree float64, memTotal float64, diskUsage float64, diskFree float64, diskTotal float64, procCount int, err error) {
	// CPU Usage usando top
	cpuCmd := "top -bn1 | grep '%Cpu' | awk '{print $2}'"
	cpuStr, err := executeSystemCommand(cpuCmd)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("error obteniendo CPU: %v", err)
	}
	cpuUsage, _ = strconv.ParseFloat(cpuStr, 64)

	// Memoria usando free
	memCmd := "free -m | grep 'Mem:'"
	memStr, err := executeSystemCommand(memCmd)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("error obteniendo memoria: %v", err)
	}
	memFields := strings.Fields(memStr)
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
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("error obteniendo disco: %v", err)
	}
	diskFields := strings.Fields(diskStr)
	if len(diskFields) >= 5 {
		diskTotal, _ = strconv.ParseFloat(strings.TrimRight(diskFields[1], "G"), 64)
		diskFree, _ = strconv.ParseFloat(strings.TrimRight(diskFields[3], "G"), 64)
		diskUsageStr := strings.TrimRight(diskFields[4], "%")
		diskUsage, _ = strconv.ParseFloat(diskUsageStr, 64)
	}

	// Número de procesos usando ps
	procCmd := "ps aux | wc -l"
	procStr, err := executeSystemCommand(procCmd)
	if err != nil {
		return cpuUsage, memUsage, memFree, memTotal, diskUsage, diskFree, diskTotal, 0, fmt.Errorf("error obteniendo procesos: %v", err)
	}
	procCount64, _ := strconv.ParseInt(procStr, 10, 64)
	procCount = int(procCount64) - 1 // Restamos 1 por el header de ps

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
