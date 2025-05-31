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
	"time"
)

// 192.168.137.102
var (
	commandMutex   sync.Mutex
	activeClients  sync.WaitGroup
	clientCount    int32
	shutdownServer = make(chan struct{})
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

	// Si es un reporte, solo procesarlo como reporte
	if strings.HasPrefix(cmd.Command, "REPORTE:") {
		fmt.Printf("Reporte recibido de %s: %s\n", cmd.User, cmd.Command)
		cmd.Response <- "Reporte recibido\n"
		return
	}

	// Si no es un reporte, ejecutar como comando
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
		if atomic.AddInt32(&clientCount, -1) == 0 {
			close(shutdownServer)
		}

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

	// Control de intentos de autenticación
	intentosRestantes := config.IntentosFallidos
	var usuario string
	var err error

	for intentosRestantes > 0 {
		// Enviar intentos restantes
		socket.Write([]byte(fmt.Sprintf("INTENTOS_RESTANTES:%d\n", intentosRestantes)))

		// Leer usuario
		usuario, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error al leer usuario: %v\n", err)
			return
		}
		usuario = strings.TrimSpace(usuario)

		// Leer contraseña
		password, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error al leer contraseña: %v\n", err)
			return
		}
		password = strings.TrimSpace(password)

		// Verificar si el usuario está en la lista de permitidos
		if !usuarioPermitido(usuario, config) {
			intentosRestantes--
			fmt.Printf("Usuario '%s' no está en la lista de permitidos. Intentos restantes: %d\n", usuario, intentosRestantes)
			socket.Write([]byte(fmt.Sprintf("AUTH_ERROR:Usuario no autorizado\n")))
			continue
		}

		// Verificar contraseña
		hash := sha256.Sum256([]byte(password))
		hashStr := hex.EncodeToString(hash[:])

		// Verificar en users.db
		autenticado := false
		file, err := os.Open("users.db")
		if err != nil {
			fmt.Printf("Error al abrir base de datos de usuarios: %v\n", err)
			socket.Write([]byte("AUTH_ERROR:Error interno del servidor\n"))
			return
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			userFile := strings.TrimSpace(parts[0])
			passFile := strings.TrimSpace(parts[1])

			if usuario == userFile && hashStr == passFile {
				autenticado = true
				break
			}
		}
		file.Close()

		if autenticado {
			socket.Write([]byte("AUTH_OK\n"))
			break
		}

		intentosRestantes--
		if intentosRestantes > 0 {
			fmt.Printf("Autenticación fallida para '%s'. Intentos restantes: %d\n", usuario, intentosRestantes)
			socket.Write([]byte("AUTH_ERROR:Contraseña incorrecta\n"))
		} else {
			fmt.Printf("Máximo de intentos alcanzado para '%s'\n", usuario)
			socket.Write([]byte("MAX_INTENTOS\n"))
			return
		}
	}

	fmt.Printf("Usuario %s autenticado desde %s\n", usuario, socket.RemoteAddr())

	// Procesar comandos del cliente
	for {
		select {
		case <-ctx.Done():
			return
		default:
			socket.SetReadDeadline(time.Now().Add(1 * time.Second))
			comando, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Printf("Cliente %s desconectado\n", socket.RemoteAddr())
					return
				}
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				fmt.Printf("Error al leer comando: %v\n", err)
				return
			}

			comando = strings.TrimSpace(comando)

			// Si el comando está vacío, ignorarlo
			if comando == "" {
				continue
			}

			// Crear el request para el comando
			cmdReq := CommandRequest{
				Command:  comando,
				User:     usuario,
				Address:  socket.RemoteAddr().String(),
				Response: make(chan string, 1),
			}

			// Procesar el comando y esperar la respuesta
			procesarComandoConcurrente(cmdReq)

			// Leer la respuesta del canal
			if respuesta, ok := <-cmdReq.Response; ok {
				// Enviar la respuesta inmediatamente al cliente
				_, err := socket.Write([]byte(respuesta))
				if err != nil {
					fmt.Printf("Error al enviar respuesta a %s: %v\n", socket.RemoteAddr(), err)
					return
				}
			}
		}
	}
}

func iniciarServidor() {
	config, err := LeerConfig("config.conf")
	if err != nil {
		fmt.Printf("Error al leer configuración: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("************************************")
	fmt.Println("*    SERVIDOR proy. oper 2025      *")
	fmt.Println("************************************")

	// Crear contexto para manejo de apagado
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar manejador de comandos
	go manejadorComandos(ctx)

	// Usar el puerto de la configuración
	socketInicial, err := net.Listen("tcp", ":"+config.Puerto)
	if err != nil {
		fmt.Printf("Error al crear socket: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Socket creado - OK en puerto", config.Puerto)
	fmt.Println("Esperando Conexiones...")
	defer socketInicial.Close()

	// Canal para señales de apagado
	go func() {
		<-shutdownServer
		fmt.Println("Iniciando apagado del servidor...")
		cancel()
		socketInicial.Close()
	}()

	for {
		socket, err := socketInicial.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Println("Error al aceptar conexión:", err)
				continue
			}
		}

		fmt.Printf("Cliente conectado desde: %s\n", socket.RemoteAddr())
		activeClients.Add(1)
		atomic.AddInt32(&clientCount, 1)
		go manejarCliente(ctx, socket, config)
	}
}
