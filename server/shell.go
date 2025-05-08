package main

import (
	"bufio"

	"fmt"

	"os"

	"os/exec"

	"strings"
)

func main() {

	fmt.Print("Digite un comando:")

	reader := bufio.NewReader(os.Stdin)

	comando, _ := reader.ReadString('\n')

	comando = strings.TrimRight(comando, "\n")

	fmt.Print("Ejecutando comando ", comando, "...\n")

	sComando := strings.Fields(comando)

	shellProy := exec.Command(sComando[0], sComando[1:]...)

	resComando, _ := shellProy.Output()

	fmt.Print("Oper#5139 ", string(resComando), "\n")
}
