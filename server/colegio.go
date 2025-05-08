package main

import "fmt"

func main() {

	var nombre string
	var edad, cantEstud, sumatoria int

	fmt.Println("Digite su capacidad de estudiantes:")
	fmt.Scanf("%d", &cantEstud)

	for i := 0; i < cantEstud; i++ {
		fmt.Println("Digite su nombre:")
		fmt.Scanf("%s", &nombre)
		fmt.Println("Digite su edad:")
		fmt.Scanf("%d", &edad)
		fmt.Println("Nuevo Programa en go desde Linux - S.O.")
		fmt.Printf("Hola %s de %d años\n", nombre, edad)
		sumatoria += edad
	}
	prom := float64(sumatoria) / float64(cantEstud)
	fmt.Println("Promedio de edad es:", prom)
}
