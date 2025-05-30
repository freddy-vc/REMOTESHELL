package main

import (
	"fmt"
	"os/exec"
	"strings"
	"runtime"
)

func ExecuteCommand(comando string) string {
	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "Error: comando vacío\n"
	}

	fmt.Printf("Ejecutando comando: %s\n", comando)

	// Determinar el shell a usar según el sistema operativo
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// En Windows, usar cmd.exe con /c para ejecutar el comando
		cmd = exec.Command("cmd.exe", "/c", comando)
	} else {
		// En sistemas Unix, usar /bin/sh
		cmd = exec.Command("/bin/sh", "-c", comando)
	}

	// Capturar tanto la salida estándar como los errores
	output, err := cmd.CombinedOutput()
	resultado := string(output)

	// Si hay error pero también hay salida, mostrar la salida
	if err != nil {
		if resultado != "" {
			return resultado
		}
		return fmt.Sprintf("Error al ejecutar comando: %v\n", err)
	}

	// Si no hay salida, enviar mensaje de confirmación
	if resultado == "" {
		return fmt.Sprintf("Comando '%s' ejecutado correctamente\n", comando)
	}

	return resultado
}
