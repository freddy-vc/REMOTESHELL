package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

func usuarioPermitido(usuario string, config *Config) bool {
	usuarios := strings.Split(config.Usuarios, ",")
	for _, u := range usuarios {
		if strings.TrimSpace(u) == strings.TrimSpace(usuario) {
			return true
		}
	}
	return false
}

func verificarIPPermitida(clienteIP string, config *Config) bool {
	return clienteIP == config.IPPermitida
}

// Verifica si un usuario existe en la base de datos
func usuarioExisteEnBD(usuario string) bool {
	file, err := os.Open("users.db")
	if err != nil {
		fmt.Printf("Error al abrir base de datos de usuarios: %v\n", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		userFile := strings.TrimSpace(parts[0])
		if usuario == userFile {
			return true
		}
	}
	return false
}

// Verifica la contraseña de un usuario
func verificarContraseña(usuario, password string) bool {
	// Calcular hash SHA256 de la contraseña
	hash := sha256.Sum256([]byte(password))
	hashStr := hex.EncodeToString(hash[:])

	// Verificar en users.db
	file, err := os.Open("users.db")
	if err != nil {
		fmt.Printf("Error al abrir base de datos de usuarios: %v\n", err)
		return false
	}
	defer file.Close()

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
			return true
		}
	}
	return false
}
