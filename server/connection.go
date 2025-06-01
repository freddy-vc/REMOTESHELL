package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	commandMutex   sync.Mutex
	activeClients  sync.WaitGroup
	clientCount    int32
	clientChannels = make(map[net.Conn]chan string)
	channelsMutex  sync.RWMutex
)

// Estructura para manejar comandos concurrentemente
type CommandRequest struct {
	Command  string
	User     string
	Address  string
	Response chan string
}

func procesarComandoConcurrente(cmd CommandRequest) {
	defer close(cmd.Response)
	commandMutex.Lock()
	respuesta := ExecuteCommand(cmd.Command, cmd.User)
	commandMutex.Unlock()
	cmd.Response <- respuesta
}

func manejadorComandos(ctx context.Context) {
	commandQueue := make(chan CommandRequest, 100)
	const maxWorkers = 5
	var wg sync.WaitGroup

	// Iniciar workers para procesar comandos
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case cmd, ok := <-commandQueue:
					if !ok {
						return
					}
					procesarComandoConcurrente(cmd)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Esperar señal de apagado
	<-ctx.Done()
	close(commandQueue)
	wg.Wait()
}

func enviarRespuestasCliente(conn net.Conn, respChan chan string) {
	for respuesta := range respChan {
		_, err := conn.Write([]byte(respuesta))
		if err != nil {
			fmt.Printf("Error al enviar respuesta a %s: %v\n", conn.RemoteAddr(), err)
			return
		}
	}
}

func manejarCliente(ctx context.Context, socket net.Conn, config *Config) {
	defer func() {
		socket.Close()
		activeClients.Done()
		atomic.AddInt32(&clientCount, -1)

		// Limpiar el canal del cliente
		channelsMutex.Lock()
		if ch, exists := clientChannels[socket]; exists {
			close(ch)
			delete(clientChannels, socket)
		}
		channelsMutex.Unlock()
	}()

	// Verificar que la IP del cliente está permitida
	clienteIP := strings.Split(socket.RemoteAddr().String(), ":")[0]
	if !verificarIPPermitida(clienteIP, config) {
		fmt.Printf("Conexión rechazada de IP no permitida: %s\n", clienteIP)
		socket.Write([]byte("IP_ERROR\n"))
		return
	}

	fmt.Printf("Cliente con IP permitida conectado: %s\n", clienteIP)
	reader := bufio.NewReader(socket)

	// Crear canal para respuestas del cliente
	respChan := make(chan string, 10)
	channelsMutex.Lock()
	clientChannels[socket] = respChan
	channelsMutex.Unlock()

	// Iniciar goroutine para enviar respuestas
	go enviarRespuestasCliente(socket, respChan)

	// Enviar número de intentos restantes
	socket.Write([]byte(fmt.Sprintf("Intentos restantes: %d\n", config.IntentosFallidos)))

	// Autenticar al usuario
	usuario, err := autenticarUsuario(reader, config)
	if err != nil {
		socket.Write([]byte("AUTH_ERROR\n"))
		return
	}

	// Enviar confirmación de autenticación exitosa
	socket.Write([]byte("AUTH_OK\n"))

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

			// Procesar el comando
			respChan := make(chan string, 1)
			req := CommandRequest{
				Command:  comando,
				User:     usuario,
				Address:  socket.RemoteAddr().String(),
				Response: respChan,
			}

			// Ejecutar el comando y enviar respuesta
			procesarComandoConcurrente(req)
			respuesta := <-respChan

			// Enviar respuesta al cliente
			_, err = socket.Write([]byte(respuesta))
			if err != nil {
				fmt.Printf("Error al enviar respuesta al cliente: %v\n", err)
				return
			}
		}
	}
}

func iniciarServidor() {
	fmt.Println("*******************************************")
	fmt.Println("*       SERVIDOR proy. oper 2025         *")
	fmt.Println("*******************************************")

	// Cargar configuración
	config, err := LeerConfig("config.conf")
	if err != nil {
		fmt.Printf("Error al cargar configuración: %v\n", err)
		os.Exit(1)
	}

	// Crear contexto para control de cierre
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar el manejador de comandos
	go manejadorComandos(ctx)

	// Iniciar el sistema de reportes
	go iniciarSistemaReportes(ctx)

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
	activeClients.Add(1)
	atomic.AddInt32(&clientCount, 1)

	// Manejar cliente en el hilo principal (no en una goroutine)
	manejarCliente(ctx, socket, config)

	fmt.Println("Servidor finalizado después de atender a un cliente")
	// El servidor termina después de atender a un cliente
}
