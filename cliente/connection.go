package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
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

// Función para obtener todas las IPs locales disponibles
func obtenerIPsLocales() ([]string, error) {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("error al obtener interfaces: %v", err)
	}

	for _, iface := range ifaces {
		// Ignorar interfaces inactivas o loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if ip4 := v.IP.To4(); ip4 != nil {
					ips = append(ips, ip4.String())
				}
			case *net.IPAddr:
				if ip4 := v.IP.To4(); ip4 != nil {
					ips = append(ips, ip4.String())
				}
			}
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no se encontraron IPs locales")
	}
	return ips, nil
}

// Función para obtener la IP local del cliente
func obtenerIPLocal() (string, error) {
	// Ejecutar el comando ipconfig
	cmd := exec.Command("ipconfig")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error al ejecutar ipconfig: %v", err)
	}

	// Convertir la salida a string y dividir por líneas
	lines := strings.Split(string(output), "\n")

	// Variables para controlar la búsqueda
	var ipWifi string
	encontradoWifi := false

	// Buscar la sección de WiFi y su IP
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Buscar el adaptador WiFi
		if strings.Contains(line, "Wi-Fi") {
			encontradoWifi = true
			continue
		}

		// Si estamos en la sección WiFi, buscar la IP
		if encontradoWifi {
			// Buscar específicamente la línea que contiene "Dirección IPv4"
			if strings.Contains(line, "IPv4") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					ip := strings.TrimSpace(parts[1])
					// Eliminar cualquier punto extra al final
					ip = strings.TrimRight(ip, ".")
					// Verificar que es una IP válida
					if net.ParseIP(ip) != nil {
						ipWifi = ip
					}
				}
			}

			// Solo terminar la sección WiFi cuando encontremos el siguiente adaptador
			// o cuando la IP ya fue encontrada
			if (strings.Contains(line, "Adaptador") && !strings.Contains(line, "Wi-Fi")) ||
				(ipWifi != "") {
				break
			}
		}
	}

	if ipWifi != "" {
		return ipWifi, nil
	}

	return "", fmt.Errorf("no se encontró la dirección IPv4 del adaptador WiFi")
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

	// Configurar un timeout para la respuesta
	socket.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer socket.SetReadDeadline(time.Time{}) // Restaurar el timeout por defecto

	// Leer respuesta del servidor
	respuesta := make([]byte, 1024)
	n, err := socket.Read(respuesta)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return fmt.Errorf("timeout esperando respuesta del servidor")
		}
		return fmt.Errorf("error al leer respuesta de autenticación: %v", err)
	}

	respuestaStr := strings.TrimSpace(string(respuesta[:n]))

	switch respuestaStr {
	case "AUTH_OK":
		return nil
	case "AUTH_ERROR":
		return fmt.Errorf("usuario no autorizado")
	case "IP_ERROR":
		return fmt.Errorf("IP no permitida")
	default:
		return fmt.Errorf("respuesta no reconocida del servidor: %s", respuestaStr)
	}
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

		fmt.Printf("\nIP local seleccionada: %s\nIP permitida: %s\n", ipLocal, ipPermitida)
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
