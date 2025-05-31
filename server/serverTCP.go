package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

// 192.168.137.102
var (
	commandMutex  sync.Mutex
	activeClients sync.WaitGroup
	clientCount   int32
)

func iniciarServidor() {
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
		activeClients.Add(1)
		atomic.AddInt32(&clientCount, 1)
		go manejarCliente(socket, config)
	}
}

func autenticarUsuario(reader *bufio.Reader, config *Config) (string, error) {
	// Leer usuario
	usuario, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer usuario: %v", err)
	}
	usuario = strings.TrimSpace(usuario)

	// Verificar si el usuario está en la lista de usuarios permitidos
	usuarioPermitido := false
	for _, u := range config.UsuariosPermitidos {
		if strings.TrimSpace(u) == usuario {
			usuarioPermitido = true
			break
		}
	}

	if !usuarioPermitido {
		return "", fmt.Errorf("usuario no autorizado")
	}

	return usuario, nil
}

func manejarCliente(socket net.Conn, config *Config) {
	defer func() {
		socket.Close()
		activeClients.Done()
		if atomic.AddInt32(&clientCount, -1) == 0 {
			fmt.Println("No hay clientes conectados. Cerrando servidor...")
			os.Exit(0)
		}
	}()

	// Verificar que la IP del cliente está permitida
	clienteIP := strings.Split(socket.RemoteAddr().String(), ":")[0]
	if clienteIP != config.IPPermitida {
		fmt.Printf("Conexión rechazada de IP no permitida: %s\n", clienteIP)
		socket.Write([]byte("IP_ERROR\n"))
		return
	}

	fmt.Printf("Cliente con IP permitida conectado: %s\n", clienteIP)
	reader := bufio.NewReader(socket)

	// Autenticar usuario
	usuario, err := autenticarUsuario(reader, config)
	if err != nil {
		fmt.Printf("Error de autenticación para cliente %s: %v\n", socket.RemoteAddr(), err)
		socket.Write([]byte("AUTH_ERROR\n"))
		return
	}

	// Enviar confirmación de autenticación exitosa
	_, err = socket.Write([]byte("AUTH_OK\n"))
	if err != nil {
		fmt.Printf("Error al enviar confirmación de autenticación: %v\n", err)
		return
	}

	fmt.Printf("Usuario %s autenticado desde %s\n", usuario, socket.RemoteAddr())

	for {
		// Leer comando del cliente
		comando, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Cliente %s desconectado\n", socket.RemoteAddr())
			} else {
				fmt.Printf("Error al leer comando: %v\n", err)
			}
			return
		}

		comando = strings.TrimSpace(comando)
		fmt.Printf("Comando recibido de %s (%s): %s\n", usuario, socket.RemoteAddr(), comando)

		// Si es un reporte, solo mostrarlo y no ejecutar como comando
		if strings.HasPrefix(comando, "__REPORTE__:") {
			fmt.Printf("Reporte recibido de %s: %s\n", usuario, comando)
			continue
		}

		// Ejecutar el comando de manera sincronizada
		commandMutex.Lock()
		respuesta := ExecuteCommand(comando)
		commandMutex.Unlock()

		// Asegurar que la respuesta termine con un salto de línea
		if !strings.HasSuffix(respuesta, "\n") {
			respuesta += "\n"
		}

		// Enviar respuesta al cliente
		_, err = socket.Write([]byte(respuesta))
		if err != nil {
			fmt.Printf("Error al enviar respuesta a %s: %v\n", socket.RemoteAddr(), err)
			return
		}
	}
}

func main() {
	iniciarServidor()
}
