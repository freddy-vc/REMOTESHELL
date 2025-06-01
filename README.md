# Sistema Cliente-Servidor para Monitoreo y Ejecución de Comandos

Este proyecto implementa un sistema cliente-servidor que permite la ejecución remota de comandos y el monitoreo de recursos del sistema. El sistema está compuesto por dos módulos principales:

## Arquitectura del Sistema

### Módulo Cliente (Windows)
- **Ubicación**: `/cliente`
- **Plataforma**: Windows
- **Funcionalidades**:
  1. Conexión remota al servidor Debian
  2. Autenticación mediante credenciales (usuario/contraseña con hash SHA-256)
  3. Envío de comandos al servidor
  4. Recepción y presentación de resultados
  5. Monitoreo periódico de recursos del servidor

### Módulo Servidor (Debian)
- **Ubicación**: `/server`
- **Plataforma**: Debian Linux
- **Funcionalidades**:
  1. Creación y gestión de socket TCP
  2. Autenticación de clientes
  3. Ejecución de comandos Unix
  4. Monitoreo de recursos del sistema
  5. Generación y envío de reportes

## Funcionalidades Detalladas

### 1. Conexión y Autenticación
- El servidor escucha conexiones TCP entrantes
- El cliente se conecta proporcionando credenciales
- Las contraseñas se transmiten hasheadas con SHA-256
- El servidor verifica las credenciales antes de aceptar la conexión

### 2. Ejecución de Comandos
- El cliente puede enviar comandos Unix al servidor
- Los comandos se ejecutan en el servidor usando `/bin/bash`
- Manejo especial del comando `cd` para navegación de directorios
- Captura y retorno de salida estándar y errores
- Sincronización mediante mutex para evitar conflictos

### 3. Monitoreo de Recursos
El servidor monitorea en tiempo real:

#### CPU
- Comando: `ps aux | awk '{sum += $3} END {print sum}'`
- Muestra: Suma total del porcentaje de CPU usado por todos los procesos activos

#### Memoria
- Comando: `free -m | grep 'Mem:'`
- Muestra:
  - Porcentaje de uso
  - Memoria libre (MB)
  - Memoria total (MB)

#### Disco
- Comando: `df -h / | tail -n 1`
- Muestra:
  - Porcentaje de uso
  - Espacio libre (GB)
  - Espacio total (GB)

#### Procesos
- Comando: `ps aux | wc -l`
- Muestra: Número total de procesos activos

### 4. Sistema de Reportes
- **Generación**: El servidor genera reportes en tiempo real
- **Periodicidad**: Configurable por el cliente
- **Formato del Reporte**:
  ```
  felipe> [DEBIAN] Recursos del Sistema:
  - CPU: XX.XX%
  - Memoria: XX.XX% (XXXX MB libre de XXXX MB)
  - Disco: XX.XX% (XX GB libre de XX GB)
  - Procesos Activos: XXX
  - Hora: YYYY-MM-DD HH:MM:SS
  ```

## Estructura del Proyecto

```
Proyecto-SO/
├── cliente/
│   ├── main.go          # Punto de entrada del cliente
│   └── report.go        # Manejo de reportes de recursos
├── server/
│   ├── main.go          # Punto de entrada del servidor
│   └── commands.go      # Ejecución de comandos y monitoreo
├── go.mod               # Dependencias del proyecto
└── go.sum               # Checksums de dependencias
```

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
