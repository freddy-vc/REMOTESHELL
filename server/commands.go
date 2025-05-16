package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func ExecuteCommand(comando string) string {
	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "Error: comando vacío\n"
	}

	fmt.Printf("Ejecutando comando: %s\n", comando)

	sComando := strings.Fields(comando)
	shellProy := exec.Command(sComando[0], sComando[1:]...)

	resComando, err := shellProy.Output()
	if err != nil {
		return fmt.Sprintf("Error al ejecutar comando: %v\n", err)
	}

	resultado := fmt.Sprintf("%s", string(resComando))
	return resultado
}
