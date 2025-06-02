package main

import (
	"fmt"
	"net"
	"strings"
)

func Conectar(ip string, puerto string, periodoReporte int) (net.Conn, string, error) {
	// Conectar al servidor
	direccion := net.JoinHostPort(ip, puerto)
	conn, err := net.Dial("tcp", direccion)
	if err != nil {
		return nil, "", fmt.Errorf("error al conectar con el servidor: %v", err)
	}

	// Leer respuesta inicial del servidor para verificar IP
	respuesta := make([]byte, 1024)
	n, err := conn.Read(respuesta)
	if err != nil {
		conn.Close()
		return nil, "", fmt.Errorf("error al leer respuesta del servidor: %v", err)
	}

	respuestaStr := strings.TrimSpace(string(respuesta[:n]))
	if respuestaStr == "IP_ERROR" {
		conn.Close()
		fmt.Println("IP NO AUTORIZADA")
		return nil, "", fmt.Errorf("IP no autorizada")
	}

	if respuestaStr != "IP_OK" {
		conn.Close()
		return nil, "", fmt.Errorf("respuesta no reconocida del servidor: %s", respuestaStr)
	}

	username, err := SolicitarCredenciales(conn)
	if err != nil {
		conn.Close()
		return nil, "", err
	}

	// Enviar periodo de reporte al servidor
	_, err = conn.Write([]byte(fmt.Sprintf("__SET_REPORT_PERIOD__%d\n", periodoReporte)))
	if err != nil {
		conn.Close()
		return nil, "", fmt.Errorf("error al enviar periodo de reporte: %v", err)
	}

	// Esperar confirmación del servidor
	respuesta = make([]byte, 1024)
	n, err = conn.Read(respuesta)
	if err != nil {
		conn.Close()
		return nil, "", fmt.Errorf("error al recibir confirmación del periodo de reporte: %v", err)
	}

	respuestaStr = strings.TrimSpace(string(respuesta[:n]))
	if respuestaStr != "REPORT_PERIOD_OK" {
		conn.Close()
		return nil, "", fmt.Errorf("error al configurar periodo de reporte: %s", respuestaStr)
	}

	return conn, username, nil
}
