package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type Config struct {
	IntentosFallidos int
	IPPermitida      string
}

// Modificamos esta función para que también lea la IP permitida
func LeerConfigIntentos(ruta string) (int, string, error) {
	file, err := os.Open(ruta)
	if err != nil {
		return 3, "", fmt.Errorf("no se pudo abrir el archivo de configuración: %v", err)
	}
	defer file.Close()

	var intentos int = 3 // Valor por defecto
	var ipPermitida string = ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "INTENTOS_MAX=") {
			fmt.Sscanf(line, "INTENTOS_MAX=%d", &intentos)
		} else if strings.HasPrefix(line, "IP_CLIENTE=") {
			ipPermitida = strings.TrimPrefix(line, "IP_CLIENTE=")
		}
	}
	return intentos, ipPermitida, nil
}

// Función para obtener la IP local del cliente
func obtenerIPLocal() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no se pudo determinar la IP local")
}

func autenticarConServidor(socket net.Conn) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese su nombre de usuario: ")
	usuario, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error al leer usuario: %v", err)
	}

	// Enviar usuario al servidor
	_, err = socket.Write([]byte(usuario))
	if err != nil {
		return fmt.Errorf("error al enviar usuario: %v", err)
	}

	// Leer respuesta del servidor
	respuesta := make([]byte, 1024)
	n, err := socket.Read(respuesta)
	if err != nil {
		return fmt.Errorf("error al leer respuesta de autenticación: %v", err)
	}

	respuestaStr := string(respuesta[:n])
	if strings.TrimSpace(respuestaStr) != "AUTH_OK" {
		return fmt.Errorf("autenticación fallida: %s", respuestaStr)
	}

	return nil
}

func Conectar(ip string, puerto string, periodoReporte int) (net.Conn, error) {
	fmt.Println("********************************")
	fmt.Println("*   CLIENTE PROY. OPER 2025    *")
	fmt.Println("********************************")

	// Leer intentos máximos e IP permitida desde el archivo de configuración del servidor
	intentosMax, ipPermitida, err := LeerConfigIntentos("../server/config.conf")
	if err != nil {
		fmt.Printf("Advertencia: %v\n", err)
	}

	// Verificar si la IP local coincide con la IP permitida
	if ipPermitida != "" {
		ipLocal, errIP := obtenerIPLocal()
		if errIP != nil {
			fmt.Printf("Error al obtener IP local: %v\n", errIP)
			return nil, errIP
		}

		fmt.Printf("IP local: %s, IP permitida: %s\n", ipLocal, ipPermitida)
		if ipLocal != ipPermitida {
			fmt.Printf("Error: La IP local (%s) no coincide con la IP permitida (%s)\n", ipLocal, ipPermitida)
			fmt.Println("Terminando el programa por seguridad...")
			os.Exit(1)
		}
	}

	var conn string = ip + ":" + puerto
	socket, err := net.Dial("tcp", conn)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar al servidor: %v", err)
	}
	fmt.Println("Conectado al socket: ", socket.RemoteAddr().String())

	// Autenticar con el servidor
	intentos := 0
	for intentos < intentosMax {
		err := autenticarConServidor(socket)
		if err == nil {
			fmt.Println("Autenticación exitosa")
			return socket, nil
		}
		fmt.Printf("Error de autenticación: %v\n", err)
		intentos++
		if intentos < intentosMax {
			fmt.Printf("Intento %d de %d. Intente nuevamente.\n", intentos+1, intentosMax)
		}
	}

	socket.Close()
	return nil, fmt.Errorf("se alcanzó el número máximo de intentos fallidos de autenticación")
}
