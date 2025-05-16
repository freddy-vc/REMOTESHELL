package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
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
	password = strings.TrimSpace(password)
	fmt.Println() // Nueva línea después de la contraseña

	if usuario == "" || password == "" {
		return "", "", fmt.Errorf("usuario y contraseña no pueden estar vacíos")
	}

	// Leer archivo de usuarios (ajusta la ruta si es necesario)
	file, err := os.Open("c:/Users/Nicolas/Documents/Proyecto-SO/server/users.db")
	if err != nil {
		return "", "", fmt.Errorf("no se pudo abrir el archivo de usuarios: %v", err)
	}
	defer file.Close()

	// Calcular hash SHA256 de la contraseña ingresada
	hash := sha256.Sum256([]byte(password))
	hashStr := hex.EncodeToString(hash[:])

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		userFile := strings.TrimSpace(parts[0])
		passFile := strings.TrimSpace(parts[1])
		if usuario == userFile && hashStr == passFile {
			return usuario, password, nil // Autenticación exitosa
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("error al leer el archivo de usuarios: %v", err)
	}

	return "", "", fmt.Errorf("usuario o contraseña incorrectos")
}
