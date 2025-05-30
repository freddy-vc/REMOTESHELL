package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Leer configuración
	config, err := LeerConfig("config.conf")
	if err != nil {
		fmt.Printf("Error al leer configuración: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("************************************")
	fmt.Println("*    SERVIDOR proy. oper 2025      *")
	fmt.Println("************************************")

	// Usar el puerto de la configuración
	socketInicial, _ := net.Listen("tcp", ":"+config.Puerto)
	fmt.Println("Socket creado - OK en puerto", config.Puerto)
	fmt.Println("Esperando Conexiones...")
	defer socketInicial.Close()

	for {
		socket, err := socketInicial.Accept()
		if err != nil {
			fmt.Println("Error al aceptar conexión:", err)
			continue
		}

		fmt.Printf("Cliente conectado desde: %s\n", socket.RemoteAddr())
		go manejarCliente(socket, config)
	}
}

func manejarCliente(socket net.Conn, config *Config) {
	defer socket.Close()

	// Verificar que la IP del cliente está permitida
	clienteIP := strings.Split(socket.RemoteAddr().String(), ":")[0]
	if clienteIP != config.IPPermitida {
		fmt.Printf("Conexión rechazada de IP no permitida: %s\n", clienteIP)
		fmt.Println("Terminando el servidor por seguridad...")
		os.Exit(1) // Termina el programa con código de error 1
	}

	fmt.Printf("Cliente con IP permitida conectado: %s\n", clienteIP)
	reader := bufio.NewReader(socket)

	for {
		// Leer comando del cliente
		comando, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Cliente %s desconectado\n", socket.RemoteAddr())
			os.Exit(0) // Cierra el servidor cuando el cliente se desconecta
		}

		comando = strings.TrimSpace(comando)
		fmt.Printf("Comando recibido de %s: %s\n", socket.RemoteAddr(), comando)

		// Si es un reporte, solo mostrarlo y no ejecutar como comando
		if strings.HasPrefix(comando, "__REPORTE__:") {
			fmt.Printf("Reporte recibido: %s\n", comando)
			continue
		}

		// Ejecutar el comando
		respuesta := ExecuteCommand(comando)

		// Enviar respuesta al cliente
		_, err = socket.Write([]byte(respuesta))
		if err != nil {
			fmt.Printf("Error al enviar respuesta a %s: %v\n", socket.RemoteAddr(), err)
			os.Exit(0) // También cierra si hay error enviando respuesta
		}
	}
}
