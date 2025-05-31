package main

import (
	"fmt"
	"os"
	"strings"
)

var nombreArchi string = "users.db"

func CrearArchivoUsuarios() {
	mensaje := []byte("felipe:caf90169eefa5f807d577486b9f795ab86ae2983c5c20806cff959117e90af18\nmaria:caf90169eefa5f807d577486b9f795ab86ae2983c5c20806cff959117e90af18\nluis:caf90169eefa5f807d577486b9f795ab86ae2983c5c20806cff959117e90af18\n")
	os.WriteFile(nombreArchi, mensaje, 0777)
}

func LeerArchivoUsuarios() {
	dbUsers, _ := os.ReadFile(nombreArchi)
	sDBusers := string(dbUsers)

	sDBusers = strings.TrimRight(sDBusers, "\n")
	cuentas := strings.Split(sDBusers, "\n")

	for _, credenciales := range cuentas {
		parametros := strings.Split(credenciales, ":")
		fmt.Println("User = ", parametros[0])
		fmt.Println("pass = ", parametros[1])
	}
}

func AgregarUsuario(username, passwordHash string) error {
	newUser := fmt.Sprintf("%s:%s\n", username, passwordHash)
	dbUsers, err := os.OpenFile(nombreArchi, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return fmt.Errorf("error al abrir archivo de usuarios: %v", err)
	}
	defer dbUsers.Close()

	_, err = dbUsers.WriteString(newUser)
	if err != nil {
		return fmt.Errorf("error al escribir nuevo usuario: %v", err)
	}

	fmt.Printf("Usuario %s agregado exitosamente\n", username)
	return nil
}
