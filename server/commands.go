package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var (
	currentDir string
	cmdMutex   sync.Mutex
)

func init() {
	// Inicializar el directorio actual
	currentDir, _ = os.Getwd()
}

func ExecuteCommand(comando string) string {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "Error: comando vacío\n"
	}

	fmt.Printf("Ejecutando comando: %s\n", comando)

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

	// Determinar el shell a usar según el sistema operativo
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// En Windows, usar la ruta completa a cmd.exe
		cmdPath := "C:\\Windows\\System32\\cmd.exe"
		cmd = exec.Command(cmdPath, "/c", comando)
	} else {
		// En sistemas Unix, usar /bin/sh
		cmd = exec.Command("/bin/sh", "-c", comando)
	}

	// Establecer el directorio de trabajo
	cmd.Dir = currentDir

	// Configurar la salida para que use los pipes correctos
	cmd.Stderr = cmd.Stdout

	// Capturar la salida
	output, err := cmd.Output()
	resultado := string(output)

	// Si no hay salida pero el comando se ejecutó correctamente
	if len(resultado) == 0 && err == nil {
		// Para comandos como 'cls' o 'clear' que no producen salida
		if comando == "cls" || comando == "clear" {
			return "\n"
		}
		return fmt.Sprintf("Comando '%s' ejecutado correctamente\n", comando)
	}

	// Si hay error pero hay salida, mostrar la salida
	if err != nil && len(resultado) > 0 {
		return resultado
	}

	// Si hay error y no hay salida, mostrar el error
	if err != nil {
		// Verificar si el error es por cmd.exe no encontrado
		if strings.Contains(err.Error(), "executable file not found") {
			// Intentar con PowerShell como alternativa
			psCmd := exec.Command("powershell.exe", "-Command", comando)
			psCmd.Dir = currentDir
			psOutput, psErr := psCmd.Output()
			if psErr == nil {
				return string(psOutput)
			}
			return fmt.Sprintf("Error al ejecutar comando (tanto en cmd como en PowerShell): %v\n", err)
		}
		return fmt.Sprintf("Error al ejecutar comando: %v\n", err)
	}

	// Asegurarse de que la salida termine con un salto de línea
	if !strings.HasSuffix(resultado, "\n") {
		resultado += "\n"
	}

	return resultado
}
