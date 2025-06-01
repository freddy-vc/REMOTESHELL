package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

func usuarioPermitido(usuario string, config *Config) bool {
	usuarios := strings.Split(config.Usuarios, ",")
	for _, u := range usuarios {
		if strings.TrimSpace(u) == strings.TrimSpace(usuario) {
			return true
		}
	}
	return false
}

func autenticarUsuario(reader *bufio.Reader, config *Config) (string, error) {
	// Leer usuario
	usuario, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer usuario: %v", err)
	}
	usuario = strings.TrimSpace(usuario)

	// Primero verificar si el usuario está en la lista de permitidos en config.conf
	if !usuarioPermitido(usuario, config) {
		fmt.Printf("Usuario '%s' no está en la lista de usuarios permitidos de config.conf\n", usuario)
		return "", fmt.Errorf("usuario no autorizado en config.conf")
	}

	// Si el usuario está permitido, verificar la contraseña
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer contraseña: %v", err)
	}
	password = strings.TrimSpace(password)

	fmt.Printf("Verificando credenciales para usuario: '%s'\n", usuario)

	// Calcular hash SHA256 de la contraseña
	hash := sha256.Sum256([]byte(password))
	hashStr := hex.EncodeToString(hash[:])

	// Verificar en users.db
	file, err := os.Open("users.db")
	if err != nil {
		return "", fmt.Errorf("error al abrir base de datos de usuarios: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		userFile := strings.TrimSpace(parts[0])
		passFile := strings.TrimSpace(parts[1])

		// Verificar si el usuario y contraseña coinciden
		if usuario == userFile && hashStr == passFile {
			fmt.Printf("Usuario '%s' autenticado exitosamente\n", usuario)
			return usuario, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error al leer base de datos de usuarios: %v", err)
	}

	fmt.Printf("Contraseña incorrecta para usuario '%s'\n", usuario)
	return "", fmt.Errorf("contraseña incorrecta")
}

func procesarComandoConcurrente(cmd CommandRequest) {
	defer close(cmd.Response)
	commandMutex.Lock()
	respuesta := ExecuteCommand(cmd.Command)
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
	if clienteIP != config.IPPermitida {
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

	// Crear socket TCP
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Puerto))
	if err != nil {
		fmt.Printf("Error al crear socket: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Servidor escuchando en el puerto %s\n", config.Puerto)

	// Aceptar conexiones
	for {
		socket, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error al aceptar conexión: %v\n", err)
			continue
		}

		activeClients.Add(1)
		atomic.AddInt32(&clientCount, 1)
		go manejarCliente(ctx, socket, config)
	}
}
