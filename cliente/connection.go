package main

import (
	"fmt"
	"net"
)

func Conectar(ip string, puerto string, periodoReporte int) (net.Conn, error) {
	fmt.Println("********************************")
	fmt.Println("*   CLIENTE PROY. OPER 2025    *")
	fmt.Println("********************************")

	var conn string = ip + ":" + puerto
	socket, err := net.Dial("tcp", conn)
	fmt.Println("Conectado al socket: ", socket.RemoteAddr().String())

	//Solicitar credenciales
	usuario, password, err := SolicitarCredenciales()
	if err != nil {
		return nil, fmt.Errorf("error al obtener credenciales: %v", err)
	}

	fmt.Printf("Usuario %s  con contraseña %s autenticado. Periodo de reporte: %d segundos\n", usuario, password, periodoReporte)

	return socket, nil
}
