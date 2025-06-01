package main

import (
	"fmt"
	"net"
)

func Conectar(ip string, puerto string, periodoReporte int) (net.Conn, string, error) {
	// Conectar al servidor
	direccion := net.JoinHostPort(ip, puerto)
	conn, err := net.Dial("tcp", direccion)
	if err != nil {
		return nil, "", fmt.Errorf("error al conectar con el servidor: %v", err)
	}

	username, err := SolicitarCredenciales(conn)
	if err != nil {
		conn.Close()
		return nil, "", err
	}

	return conn, username, nil
}
