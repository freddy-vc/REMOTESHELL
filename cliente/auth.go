package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func SolicitarCredenciales() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Solicitar usuario
	fmt.Print("Usuario: ")
	usuario, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	usuario = strings.TrimSpace(usuario)

	// Solicitar contraseña
	fmt.Print("Contraseña: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Println() // Nueva línea después de la contraseña

	if usuario == "" || password == "" {
		return "", "", fmt.Errorf("usuario y contraseña no pueden estar vacíos")
	}

	return usuario, password, nil
}
