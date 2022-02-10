package cpu

import (
	"THUMB/memory"
	"fmt"
	"strings"
)

type PSR struct {
	Status string
	Mode   string
}
type Instruction struct {
	Value   int
	Address int
}

func isUpper(char byte) bool {
	if char <= 'z' && char >= 'a' {
		return false
	} else {
		return true
	}
}
func STRH(address int, value int16) bool {
	return STR(address, int(value))
}
func STRB(address int, value int8) bool {
	return STR(address, int(value))
}

func STR(address int, value int) bool {
	if address > 0x80000000 && address <= 0x80000060 {
		return false
	}
	for i := range memory.RAM {
		if memory.RAM[i].Address == address {
			memory.RAM[i].Value = value
			return true
		}
	}
	memory.RAM = append(memory.RAM, memory.Data{
		Address: address,
		Value:   value,
	})
	return true
}

func LDR(address int) uint {
	for i := range memory.RAM {
		if memory.RAM[i].Address == address {
			return uint(memory.RAM[i].Value)
		}
	}
	return 1 //LIXO
}
func LDRB(address int) uint8 {
	for i := range memory.RAM {
		if memory.RAM[i].Address == address {
			return uint8(memory.RAM[i].Value)
		}
	}
	return 1 //LIXO
}
func LDRH(address int) uint16 {
	for i := range memory.RAM {
		if memory.RAM[i].Address == address {
			return uint16(memory.RAM[i].Value)
		}
	}
	return 1 //LIXO
}
func LDRSB(address int) int8 {
	for i := range memory.RAM {
		if memory.RAM[i].Address == address {
			return int8(memory.RAM[i].Value)
		}
	}
	return -1 //LIXO
}

func LDRSH(address int) int16 {
	for i := range memory.RAM {
		if memory.RAM[i].Address == address {
			return int16(memory.RAM[i].Value)
		}
	}
	return -1 //LIXO
}
func Push(value int) {
	for i := range memory.RAM {
		if memory.RAM[i].Address == SP {
			memory.RAM[i].Value = value
			SP += 4
			return
		}
	}
	memory.RAM = append(memory.RAM, memory.Data{
		Address: SP,
		Value:   value,
	})
	SP += 4
}
func Pop() int {
	for i := range memory.RAM {
		if memory.RAM[i].Address == SP-4 {
			SP -= 4
			return memory.RAM[i].Value
		}
	}

	fmt.Println("Data Abort")

	return 0

}

func SetCarryAndOverflow(a int, b int, tp string) {
	switch tp {
	case "+":
		if uint64(a)+uint64(b) == uint64(uint32(a)+uint32(b)) {
			UpdateCPSR("C", "c")
		} else {
			UpdateCPSR("c", "C")
		}

		if int64(a)+int64(b) == int64(int32(a)+int32(b)) {
			UpdateCPSR("V", "v")
		} else {
			UpdateCPSR("v", "V")
		}
	case "-":
		if a >= b {
			UpdateCPSR("c", "C")
		} else if uint64(a)-uint64(b) == uint64(uint32(a)-uint32(b)) {
			UpdateCPSR("C", "c")
		} else {
			UpdateCPSR("c", "C")
		}
		if a == 0 && b == 0 || int64(a)-int64(b) == int64(int32(a)-int32(b)) {
			UpdateCPSR("V", "v")
		} else {
			UpdateCPSR("v", "V")
		}
	case ">>":
		if a != 0 {
			if uint64(a)<<uint64(b) == uint64(uint32(a)<<uint32(b)) {
				UpdateCPSR("C", "c")
			} else {
				UpdateCPSR("c", "C")
			}
		}

	}
}
func NegativeOrZero(tmp int) {
	if tmp == 0 {
		UpdateCPSR("z", "Z")
		UpdateCPSR("N", "n")
	} else if tmp < 0 {
		UpdateCPSR("Z", "z")
		UpdateCPSR("n", "N")
	} else {
		UpdateCPSR("Z", "z")
		UpdateCPSR("N", "n")
	}
}

func UpdateCPSR(old string, new string) {
	CPSR.Status = strings.Replace(CPSR.Status, old, new, 1)
}

func PrintState() {
	fmt.Println("--------------Registradores---------------")
	for i := range R {
		fmt.Printf("R%d : 0x%X \n", i, R[i])
	}
	fmt.Printf("PC : 0x%X SP : 0x%X  LR : 0x%X\n", PC, SP, LR)
	fmt.Println("---------------Status---------------")
	fmt.Println(CPSR.Status + "_" + CPSR.Mode)
	if isUpper(CPSR.Status[8]) {
		fmt.Println("Mode Thumb")
	}
}
func PrintInstructions() {
	fmt.Println("----------- INSTRUÇÕES -----------")
	for k := range INSTRUCTIONS {
		fmt.Printf("Endereço: 0x%X Valor: 0x%X\n", INSTRUCTIONS[k].Address, INSTRUCTIONS[k].Value)
	}
	fmt.Println("------------------------------")
}
