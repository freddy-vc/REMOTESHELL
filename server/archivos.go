package main

import (
	"fmt"
	"os"
	"strings"
)

var nombreArchi string = "users.db"

func main() {
	//crearArchi()
	leerArchi()
	modificarArchi()
}

func crearArchi() {
	mensaje := []byte("felipe:caf90169eefa5f807d577486b9f795ab86ae2983c5c20806cff959117e90af18\nmaria:caf90169eefa5f807d577486b9f795ab86ae2983c5c20806cff959117e90af18\nluis:caf90169eefa5f807d577486b9f795ab86ae2983c5c20806cff959117e90af18\n")

	os.WriteFile(nombreArchi, mensaje, 0777)
}

func leerArchi() {

	dbUsers, _ := os.ReadFile(nombreArchi)
	sDBusers := string(dbUsers)
	//fmt.Println(sDBusers)

	sDBusers = strings.TrimRight(sDBusers, "\n")
	cuentas := strings.Split(sDBusers, "\n")

	for _, credenciales := range cuentas {
		//separar parametros de la cuenta
		parametros := strings.Split(credenciales, ":")
		fmt.Println("User = ", parametros[0])
		fmt.Println("pass = ", parametros[1])
	}
}

func modificarArchi() {
	fmt.Println("...Agregando usuario...")
	newUser := "jose:caf90169eefa5f807d577486b9f795ab86ae2983c5c20806cff959117e90af18\n"
	dbUsers, _ := os.OpenFile(nombreArchi, os.O_APPEND|os.O_WRONLY, 0777)
	dbUsers.WriteString(newUser)
	dbUsers.Close()
	fmt.Println("...Nuevo contenido...")

	leerArchi()
}
