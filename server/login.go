package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	//passBD es "linux"
	const passwordBD string = "caf90169eefa5f807d577486b9f795ab86ae2983c5c20806cff959117e90af18"

	fmt.Println("*******************************************")
	fmt.Println("*       PROYECTO OEPRATIVOS 2025-1        *")
	fmt.Println("*******************************************\n")
	var passUS string
	fmt.Print("Digite Password:")
	fmt.Scanf("%s", &passUS)
	fmt.Println("Password es:", passUS)

	hashUS := sha256.Sum256([]byte(passUS))
	HexUS := fmt.Sprintf("%x", hashUS)

	if passwordBD == HexUS {
		fmt.Print("---Login exitoso!---\n")
	} else {
		fmt.Print("---Login incorrecto!---\n")
	}

}
