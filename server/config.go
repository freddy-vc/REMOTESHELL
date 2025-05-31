package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	IPPermitida        string
	Puerto             string
	IntentosFallidos   int
	UsuariosPermitidos []string
	Usuarios           string
}

// Lee el archivo de configuración y retorna una estructura Config
func LeerConfig(ruta string) (*Config, error) {
	file, err := os.Open(ruta)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el archivo de configuración: %v", err)
	}
	defer file.Close()

	config := &Config{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Ignorar líneas vacías o comentarios
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "IP_CLIENTE":
			config.IPPermitida = value
		case "PUERTO":
			config.Puerto = value
		case "INTENTOS_MAX":
			config.IntentosFallidos = 3 // valor por defecto
			fmt.Sscanf(value, "%d", &config.IntentosFallidos)
		case "USUARIOS":
			config.Usuarios = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error al leer el archivo de configuración: %v", err)
	}
	return config, nil
}
