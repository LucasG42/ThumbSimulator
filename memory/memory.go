package memory

import (
	"THUMB/input"
	"fmt"
	"strconv"
	"strings"
)

type Data struct {
	Address int
	Value   int
}

var RAM = make([]Data, 0)

const offset = 0x80000000

func DataMem() {
	for i := range RAM {
		if RAM[i].Address > 0x80000060 || RAM[i].Address < 0x80000000 {
			fmt.Printf("Endereço: 0x%X Dado: 0x%X\n", RAM[i].Address, RAM[i].Value)
		}
	}

}

func Init() bool {

	tmp := strings.Split(input.Buffer, "\n")

	for i := range tmp {
		address, err := strconv.ParseInt(strings.Split(tmp[i], " ")[0], 16, 64)
		value, err := strconv.ParseInt(strings.Split(tmp[i], ": ")[1], 16, 64)
		if err == nil && address <= 0x60 {
			RAM = append(RAM, Data{
				Address: int(address + offset),
				Value:   int(value),
			})
		} else {
			return false
		}
	}
	return true
}
func PrintRAM() {
	for i := range RAM {
		fmt.Printf("Endereço: %X Dado: %X\n", RAM[i].Address, RAM[i].Value)
	}
}
