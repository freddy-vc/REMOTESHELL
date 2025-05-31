package main

import (
	"bufio"
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

// 192.168.137.102
var (
	commandMutex  sync.Mutex
	activeClients sync.WaitGroup
	clientCount   int32
)

func autenticarUsuario(reader *bufio.Reader, config *Config) (string, error) {
	// Leer usuario
	usuario, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer usuario: %v", err)
	}
	usuario = strings.TrimSpace(usuario)

	// Leer contraseña
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer contraseña: %v", err)
	}
	password = strings.TrimSpace(password)

	fmt.Printf("Intento de autenticación del usuario: '%s'\n", usuario)

	// Calcular hash SHA256 de la contraseña
	hash := sha256.Sum256([]byte(password))
	hashStr := hex.EncodeToString(hash[:])

	// Leer archivo de usuarios y contraseñas
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

	fmt.Printf("Usuario '%s' no autorizado\n", usuario)
	return "", fmt.Errorf("usuario o contraseña incorrectos")
}

func procesarComando(comando string, usuario string, addr string) string {
	comando = strings.TrimSpace(comando)
	fmt.Printf("Comando recibido de %s (%s): %s\n", usuario, addr, comando)

	// Si es un reporte, solo mostrarlo y no ejecutar como comando
	if strings.HasPrefix(comando, "__REPORTE__:") {
		fmt.Printf("Reporte recibido de %s: %s\n", usuario, comando)
		return "Reporte recibido\n"
	}

	// Ejecutar el comando de manera sincronizada
	commandMutex.Lock()
	respuesta := ExecuteCommand(comando)
	commandMutex.Unlock()

	return respuesta
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

	// Buffer para leer comandos
	cmdBuffer := make([]byte, 4096)

	for {
		// Leer comando del cliente
		n, err := reader.Read(cmdBuffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Cliente %s desconectado\n", socket.RemoteAddr())
			} else {
				fmt.Printf("Error al leer comando: %v\n", err)
			}
			return
		}

		comando := string(cmdBuffer[:n])
		respuesta := procesarComando(comando, usuario, socket.RemoteAddr().String())

		// Enviar respuesta al cliente
		_, err = socket.Write([]byte(respuesta))
		if err != nil {
			fmt.Printf("Error al enviar respuesta a %s: %v\n", socket.RemoteAddr(), err)
			return
		}
	}
}

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

func main() {
	iniciarServidor()
}
