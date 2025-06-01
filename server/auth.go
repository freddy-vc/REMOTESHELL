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

func autenticarUsuario(reader *bufio.Reader, config *Config) (string, error) {
	// Leer usuario
	usuario, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer usuario: %v", err)
	}
	usuario = strings.TrimSpace(usuario)

	// Primero verificar si el usuario está en la lista de permitidos en config.conf
	if !usuarioPermitido(usuario, config) {
		fmt.Printf("Usuario '%s' no está en la lista de usuarios permitidos de config.conf\n", usuario)
		return "", fmt.Errorf("usuario no autorizado en config.conf")
	}

	// Si el usuario está permitido, verificar la contraseña
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer contraseña: %v", err)
	}
	password = strings.TrimSpace(password)

	fmt.Printf("Verificando credenciales para usuario: '%s'\n", usuario)

	// Verificar si el usuario existe en users.db
	usuarioExiste := false

	// Calcular hash SHA256 de la contraseña
	hash := sha256.Sum256([]byte(password))
	hashStr := hex.EncodeToString(hash[:])

	// Verificar en users.db
	file, err := os.Open("users.db")
	if err != nil {
		return "", fmt.Errorf("error al abrir base de datos de usuarios: %v", err)
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

		// Verificar si el usuario existe
		if usuario == userFile {
			usuarioExiste = true
			// Verificar si la contraseña coincide
			if hashStr == passFile {
				fmt.Printf("Usuario '%s' autenticado exitosamente\n", usuario)
				return usuario, nil
			} else {
				fmt.Printf("Contraseña incorrecta para usuario '%s'\n", usuario)
				return "", fmt.Errorf("contraseña incorrecta")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error al leer base de datos de usuarios: %v", err)
	}

	if !usuarioExiste {
		fmt.Printf("Usuario '%s' no encontrado en la base de datos\n", usuario)
		return "", fmt.Errorf("usuario no encontrado")
	}

	// Este código no debería ejecutarse nunca, pero lo dejamos por seguridad
	return "", fmt.Errorf("error de autenticación")
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
