package main

import (
	"THUMB/cpu"
	"THUMB/input"
	"THUMB/memory"
	"fmt"
	"os"
)

func main() {
	if input.LoadFile(os.Args[1]) {
		fmt.Println("Arquivo carregado com sucesso")
		if memory.Init() == false {
			fmt.Println("Falha ao ler arquivo")
			return
		} else {
			fmt.Println("--------------RAM--------------")
			memory.PrintRAM()
		}
		cpu.Init()

		for {
			if cpu.Cycle() == false {
				cpu.PrintState()
				fmt.Println("--------- Memória de Dados  -------")
				memory.DataMem()
				break
			}
		}

	} else {
		fmt.Println("Falha ao carregar")
	}
}
