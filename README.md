# Sistema Cliente-Servidor para Ejecución Remota de Comandos

Este proyecto implementa un sistema cliente-servidor que permite la ejecución remota de comandos UNIX y el monitoreo de recursos del sistema.

## Módulo Linux (Servidor)

1. Abre un socket Servidor y espera conexión remota.
2. Lee archivo plano de configuración (.conf) donde obtiene:
   - IP permitida al cliente
   - Puerto por el que atiende
   - Cantidad de intentos fallidos de autenticación
   - Usuarios permitidos

3. Autentica usuario remoto contra una base de datos de usuario (archivo plano).
   La validación del password se hace con el algoritmo de hash SHA-256.

4. En paralelo:
   - Recibe comando remoto, lo ejecuta y envía la respuesta por el socket.
   - Genera reporte de consumo de recursos (procesador, procesos, memoria y disco).
     Luego lo envía por socket cada n segundos (n es un parámetro recibido remotamente).

## Módulo Windows (Cliente)

1. Se conecta al Servidor remoto (recibe parámetros de conexión al momento de ejecutar la aplicación):
   - IP
   - Puerto
   - PeriodoReporte

   Ejemplo: `clienteOperativo 10.1.3.3 2306 5`
   (Se va a conectar a la IP 10.1.3.3, por el puerto 2306 y va a recibir reportes cada 5 segundos)

2. Solicita al usuario credenciales de acceso (Usuario y password)

3. En paralelo:
   - Permite la ejecución de comandos UNIX remotamente.
   - Presenta reporte de consumo de recursos cada n segundos.

## Estructura del Proyecto

```
Proyecto-SO/
├── cliente/
│   ├── main.go          # Punto de entrada y manejo de parámetros
│   ├── connection.go    # Conexión al servidor
│   ├── auth.go         # Autenticación
│   ├── commands.go     # Ejecución de comandos
│   └── report.go       # Presentación de reportes
├── server/
│   ├── main.go         # Punto de entrada
│   ├── serverTCP.go    # Socket y manejo de conexiones
│   ├── commands.go     # Ejecución de comandos y monitoreo
│   ├── config.go       # Lectura de configuración
│   ├── config.conf     # Archivo de configuración
│   └── users.db        # Base de datos de usuarios
└── go.mod              # Dependencias del proyecto

## Monitoreo de Recursos

El servidor monitorea en tiempo real:

### CPU
- Comando: `vmstat 1 2 | tail -1 | awk '{print 100 - $15 "%"}'`
- Muestra: Porcentaje de uso de CPU basado en el tiempo idle del sistema

### Memoria
- Comando: `free -m | grep 'Mem:'`
- Muestra:
  - Porcentaje de uso
  - Memoria libre (MB)
  - Memoria total (MB)

### Disco
- Comando: `df -h / | tail -n 1`
- Muestra:
  - Porcentaje de uso
  - Espacio libre (GB)
  - Espacio total (GB)

### Procesos
- Comando: `ps aux | wc -l`
- Muestra: Número total de procesos activos

## Protocolo de Comunicación

### Comandos Especiales
1. `__GET_REPORT__`: Solicita un reporte de recursos
2. `__SYNC__`: Sincronización de comandos
3. `cd`: Manejo especial para cambio de directorio

### Flujo de Datos
1. **Ejecución de Comandos**:
   ```
   Cliente -> Servidor: [COMANDO]
   Servidor -> Cliente: [RESULTADO]
   ```

2. **Monitoreo de Recursos**:
   ```
   Cliente -> Servidor: __GET_REPORT__
   Servidor: [Ejecuta comandos de monitoreo]
   Servidor -> Cliente: [REPORTE FORMATEADO]
   ```

## Características de Seguridad
1. Autenticación mediante hashing SHA-256
2. Mutex para sincronización de comandos
3. Manejo de errores y timeouts
4. Limpieza de recursos al desconectar

## Requisitos del Sistema
- **Cliente**: Windows 10 o superior
- **Servidor**: Debian Linux
- **Go**: versión 1.16 o superior

## Configuración y Ejecución

### Servidor (Debian)
1. Navegar al directorio del servidor:
   ```bash
   cd server
   ```
2. Ejecutar el servidor:
   ```bash
   go run .
   ```

### Cliente (Windows)
1. Navegar al directorio del cliente:
   ```bash
   cd cliente
   ```
2. Ejecutar el cliente:
   ```bash
   go run .
   ```

## Manejo de Errores
- Reconexión automática en caso de pérdida de conexión
- Timeout configurable para comandos
- Logs detallados de errores
- Recuperación de estados inconsistentes

## Limitaciones
- Solo soporta comandos Unix en el servidor
- Requiere permisos de administrador para algunos comandos
- La autenticación es básica (no usa certificados SSL/TLS)
