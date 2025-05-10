package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Obtener parámetros de conexión
	ip, puerto, periodoReporte, err := SolicitarParametros()
	if err != nil {
		fmt.Printf("Error al obtener parámetros: %v\n", err)
		os.Exit(1)
	}

	// Intentar conexión con el servidor
	err = Conectar(ip, puerto, periodoReporte)
	if err != nil {
		fmt.Printf("Error al conectar con el servidor: %v\n", err)
		os.Exit(1)
	}
}

func SolicitarParametros() (string, string, int, error) {
	reader := bufio.NewReader(os.Stdin)

	// Solicitar IP
	fmt.Print("Ingrese la IP del servidor: ")
	ip, err := reader.ReadString('\n')
	if err != nil {
		return "", "", 0, err
	}
	ip = strings.TrimSpace(ip)

	// Solicitar Puerto
	fmt.Print("Ingrese el puerto del servidor: ")
	puerto, err := reader.ReadString('\n')
	if err != nil {
		return "", "", 0, err
	}
	puerto = strings.TrimSpace(puerto)

	// Solicitar Periodo de Reporte
	fmt.Print("Ingrese el periodo de reporte en segundos: ")
	periodoStr, err := reader.ReadString('\n')
	if err != nil {
		return "", "", 0, err
	}
	periodoStr = strings.TrimSpace(periodoStr)

	// Convertir periodo a entero
	periodo, err := strconv.Atoi(periodoStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("el periodo debe ser un número válido")
	}

	// Validar parámetros
	if ip == "" || puerto == "" || periodo <= 0 {
		return "", "", 0, fmt.Errorf("todos los parámetros son obligatorios y el periodo debe ser mayor a 0")
	}

	return ip, puerto, periodo, nil
}
