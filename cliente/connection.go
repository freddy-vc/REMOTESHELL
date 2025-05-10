package main

import (
	"fmt"
	"net"
)

func Conectar(ip string, puerto string, periodoReporte int) error {
	fmt.Println("********************************")
	fmt.Println("*   CLIENTE PROY. OPER 2025    *")
	fmt.Println("********************************")

	var conn string = ip + ":" + puerto
	socket, _ := net.Dial("tcp", conn)
	fmt.Println("Conectado al socket: ", socket.RemoteAddr().String())

	//Solicitar credenciales
	usuario, password, err := SolicitarCredenciales()
	if err != nil {
		return fmt.Errorf("error al obtener credenciales: %v", err)
	}

	fmt.Printf("Usuario %s  con contraseña %s autenticado. Periodo de reporte: %d segundos\n", usuario, password, periodoReporte)

	return nil
}
