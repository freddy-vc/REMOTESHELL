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

	// Manejar reportes del cliente
	if strings.HasPrefix(comando, "__REPORTE__:") {
		fmt.Println(comando) // Imprimir el reporte en el servidor
		return "Reporte recibido\n"
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
