package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	currentDir string
	cmdMutex   sync.Mutex
)

func ExecuteCommand(comando string, username string) string {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "\n"
	}

	// Manejar solicitud de reporte
	if comando == "__GET_REPORT__" {
		fmt.Println("Recibida solicitud de reporte")
		return generateSystemReport(username)
	}

	fmt.Printf("Ejecutando comando UNIX: %s\n", comando)

	// Manejar el comando cd de forma especial
	if strings.HasPrefix(comando, "cd ") {
		dir := strings.TrimPrefix(comando, "cd ")
		dir = strings.TrimSpace(dir)

		err := os.Chdir(dir)
		if err != nil {
			return fmt.Sprintf("Error al cambiar al directorio %s: %v\n", dir, err)
		}

		currentDir, _ = os.Getwd()
		return fmt.Sprintf("Directorio cambiado a: %s\n", currentDir)
	}

	// Ejecutar comando usando /bin/bash
	cmd := exec.Command("/bin/bash", "-c", comando)
	cmd.Dir = currentDir

	output, err := cmd.CombinedOutput()
	resultado := string(output)

	if err != nil {
		if len(resultado) > 0 {
			return resultado
		}
		return fmt.Sprintf("Error al ejecutar comando: %v\n", err)
	}

	if strings.TrimSpace(resultado) == "" {
		return fmt.Sprintf("Comando '%s' ejecutado correctamente\n", comando)
	}

	if !strings.HasSuffix(resultado, "\n") {
		resultado += "\n"
	}

	return resultado
}
