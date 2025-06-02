package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var commandMutex sync.Mutex

func manejarCliente(ctx context.Context, socket net.Conn, config *Config) {
	defer socket.Close()

	// Verificar que la IP del cliente está permitida
	clienteIP := strings.Split(socket.RemoteAddr().String(), ":")[0]
	if !verificarIPPermitida(clienteIP, config) {
		fmt.Printf("Conexión rechazada de IP no permitida: %s\n", clienteIP)
		socket.Write([]byte("IP_ERROR\n"))
		return
	}

	fmt.Printf("Cliente con IP permitida conectado: %s\n", clienteIP)
	socket.Write([]byte("IP_OK\n"))

	reader := bufio.NewReader(socket)

	// Proceso de autenticación con intentos
	intentos := config.IntentosFallidos
	var usuario string

	for intentos > 0 {
		// Enviar número de intentos restantes
		socket.Write([]byte(fmt.Sprintf("Intentos restantes: %d\n", intentos)))

		// Solicitar usuario
		usuario, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error al leer usuario: %v\n", err)
			return
		}
		usuario = strings.TrimSpace(usuario)

		// Verificar si el usuario existe en la base de datos
		if !usuarioExisteEnBD(usuario) {
			fmt.Printf("Usuario '%s' no encontrado en la base de datos\n", usuario)
			socket.Write([]byte("USER_NOT_FOUND\n"))
			intentos--
			continue
		}

		// Verificar si el usuario está permitido en config.conf
		if !usuarioPermitido(usuario, config) {
			fmt.Printf("Usuario '%s' no está en la lista de usuarios permitidos de config.conf\n", usuario)
			socket.Write([]byte("USER_NOT_ALLOWED\n"))
			intentos--
			continue
		}

		// Usuario válido y permitido, enviar confirmación
		socket.Write([]byte("USER_VALID\n"))

		// Solicitar contraseña
		password, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error al leer contraseña: %v\n", err)
			return
		}
		password = strings.TrimSpace(password)

		// Verificar contraseña
		if verificarContraseña(usuario, password) {
			fmt.Printf("Usuario '%s' autenticado exitosamente\n", usuario)
			socket.Write([]byte("AUTH_OK\n"))
			break // Autenticación exitosa, salir del bucle
		} else {
			fmt.Printf("Contraseña incorrecta para usuario '%s'\n", usuario)
			socket.Write([]byte("PASSWORD_ERROR\n"))
			intentos--
		}
	}

	// Verificar si se agotaron los intentos
	if intentos <= 0 {
		fmt.Printf("Se agotaron los intentos para el cliente %s\n", clienteIP)
		socket.Write([]byte("MAX_ATTEMPTS\n"))
		return
	}

	// Procesar comandos del cliente
	for {
		select {
		case <-ctx.Done():
			return
		default:
			comando, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error al leer comando del cliente: %v\n", err)
				}
				return
			}

			comando = strings.TrimSpace(comando)
			if comando == "" {
				continue
			}

			// Verificar si es comando de configuración de periodo de reporte
			if strings.HasPrefix(comando, "__SET_REPORT_PERIOD__") {
				periodoStr := strings.TrimPrefix(comando, "__SET_REPORT_PERIOD__")
				periodo, err := strconv.Atoi(periodoStr)
				if err != nil {
					socket.Write([]byte("REPORT_PERIOD_ERROR\n"))
					continue
				}
				// Iniciar el sistema de reportes con el periodo especificado
				go iniciarSistemaReportes(ctx, socket, usuario, periodo)
				socket.Write([]byte("REPORT_PERIOD_OK\n"))
				continue
			}

			// Procesar otros comandos
			commandMutex.Lock()
			respuesta := ExecuteCommand(comando, usuario)
			commandMutex.Unlock()

			_, err = socket.Write([]byte(respuesta))
			if err != nil {
				fmt.Printf("Error al enviar respuesta: %v\n", err)
				return
			}
		}
	}
}

func iniciarServidor() {
	fmt.Println("[x]----------|||---------[x]")
	fmt.Println("|   REMOTESHELL - SERVER   |")
	fmt.Println("[x]----------|||---------[x]")

	// Cargar configuración
	config, err := LeerConfig("config.conf")
	if err != nil {
		fmt.Printf("Error al cargar configuración: %v\n", err)
		os.Exit(1)
	}

	// Crear contexto para control de cierre
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Crear socket TCP
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Puerto))
	if err != nil {
		fmt.Printf("Error al crear socket: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Servidor escuchando en el puerto %s\n", config.Puerto)
	fmt.Printf("Esperando conexiones...\n")

	// Aceptar una única conexión
	fmt.Println("El servidor solo aceptará una conexión")
	socket, err := listener.Accept()
	if err != nil {
		fmt.Printf("Error al aceptar conexión: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Conexión aceptada desde: %s\n", socket.RemoteAddr().String())

	// Manejar cliente en el hilo principal
	manejarCliente(ctx, socket, config)

	fmt.Println("Servidor finalizado después de atender a un cliente")
}
