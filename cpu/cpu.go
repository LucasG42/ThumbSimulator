package cpu

import (
	"THUMB/memory"
	"fmt"
)

var INSTRUCTIONS = make([]Instruction, 0)

var R [8]int
var PC int
var LR int
var SP int
var CPSR PSR

func Init() {
	PC = 0x80000000
	for k := range memory.RAM {
		INSTRUCTIONS = append(INSTRUCTIONS, Instruction{Value: memory.RAM[k].Value & 0x0000FFFF, Address: PC})

		INSTRUCTIONS = append(INSTRUCTIONS, Instruction{Value: memory.RAM[k].Value & 0xFFFF0000 >> 16, Address: PC + 2})
		PC += 4
	}
	SP = 0x82000000
	PC = 0x80000000
	CPSR.Status = "nzcvqjifT"
	CPSR.Mode = "SVC"
}

func Cycle() bool {
	for k := range INSTRUCTIONS {
		if INSTRUCTIONS[k].Address == PC {
			if INSTRUCTIONS[k].Value == 0x00000000 {
				return false
			}
			if Decode(INSTRUCTIONS[k]) == false || INSTRUCTIONS[k].Value == 0xE7FE {
				return false
			}
		}
	}
	return true
}
func Decode(instruction Instruction) bool {
	PC += 2
	switch (instruction.Value & 0xF000) >> 12 {
	// BIT 15 - 12
	//
	case 0x0:
		if (instruction.Value&0x0800)>>11 == 0 {
			// LSL
			Lm := (instruction.Value & 0x0038) >> 3
			tmp := (instruction.Value & 0x07C0) >> 6
			R[instruction.Value&0x0007] = R[Lm] << tmp
		} else {
			// LSR
			Lm := (instruction.Value & 0x0038) >> 3
			tmp := (instruction.Value & 0x07C0) >> 6
			R[instruction.Value&0x0007] = R[Lm] >> tmp
		}
	case 0x1:
		//BIT 11
		if (instruction.Value&0x0800)>>11 == 0 {
			// ASR
		} else {
			// ADD | SUB
			// BIT 10 - 9
			//TMP = Immed3 || LM

			tmp := (instruction.Value & 0x01C0) >> 6 //BIT 8 - 6
			ln := instruction.Value & 0x0038 >> 3    //BIT 5 - 3
			ld := instruction.Value & 0x0007         //BIT 2 - 0
			// BIT 10 - 9
			switch (instruction.Value & 0x0600) >> 9 {
			case 0x0:
				// ADD 3 OPERANDOS
				R[ld] = R[ln] + R[tmp]
			case 0x1:
				// SUB 3 OPERANDOS
				R[ld] = R[ln] - R[tmp]
			case 0x2:
				R[ld] = R[ln] + tmp
				// ADD 2 OPERANDOS
			case 0x3:
				// SUB  2 OPERANDOS
				R[ld] = R[ln] - tmp
			}
		}
	case 0x2:
		// OP
		if (instruction.Value&0x0800)>>11 == 0 {
			// MOV
			// R[BIT 10 - 8] = BIT 7 - 0
			R[(instruction.Value&0x0700)>>8] = instruction.Value & 0x00FF
			NegativeOrZero(R[(instruction.Value&0x0700)>>8])
		} else {
			//CPM
			SetCarryAndOverflow(R[(instruction.Value&0x0700)>>8], instruction.Value&0x00FF, "-")
			NegativeOrZero(R[(instruction.Value&0x0700)>>8] - instruction.Value&0x00FF)
		}
	case 0x3:
		if (instruction.Value&0x0800)>>11 == 0 {
			R[(instruction.Value&0x0700)>>8] += instruction.Value & 0x00FF
			// ADD 1 OPERANDO
		} else {
			R[(instruction.Value&0x0700)>>8] -= instruction.Value & 0x00FF
			// SUB 1 OPERANDO
		}
	case 0x4:
		// BIT 11
		if (instruction.Value&0x0800)>>11 == 0 {
			// BIT 10 - 8
			lm := instruction.Value & 0x0038 >> 3 //BIT 5 - 3
			ld := instruction.Value & 0x0007      //BIT 2 - 0
			switch (instruction.Value & 0x0700) >> 8 {
			case 0x0:
				// BIT 7 - 6
				switch (instruction.Value & 0x0C00) >> 6 {
				case 0x0:
					R[ld] &= R[lm]
				case 0x1:
					R[ld] ^= R[lm]
				case 0x2:
					R[ld] = R[ld] << R[lm]
				case 0x3:
					R[ld] = R[ld] >> R[lm]
				}
			case 0x1:
				carry := 0
				if isUpper(CPSR.Status[2]) {
					carry = 1
				}
				switch (instruction.Value & 0x0C00) >> 6 {
				case 0x0:
					// ASR
				case 0x1:
					R[ld] += R[lm] + carry
				case 0x2:
					R[ld] = R[ld] - R[lm] - carry
				case 0x3:
					// ROR
					for j := 0; j < R[lm]; j++ {
						if R[ld]&0x00000001 == 1 {
							UpdateCPSR("c", "C")
							R[ld] = R[ld] >> 1
							R[ld] |= 0x80000000
						} else {
							UpdateCPSR("C", "c")
							R[ld] = R[ld] >> 1
						}
					}
				}
			case 0x2:
				switch (instruction.Value & 0x0C00) >> 6 {
				case 0x0:
					// TST
					tmp := R[lm] & R[ld]
					NegativeOrZero(tmp)
				case 0x1:
					// NEG
					R[ld] = 0 - R[lm]
				case 0x2:
					// CMP
					SetCarryAndOverflow(R[ld], R[lm], "-")
					NegativeOrZero(R[ld] - R[lm])
				case 0x3:
					// CMN
					SetCarryAndOverflow(R[ld], R[lm], "+")
					NegativeOrZero(R[ld] + R[lm])
				}
			case 0x3:
				switch (instruction.Value & 0x0C00) >> 6 {
				case 0x0:
					//ORR
					R[ld] |= R[lm]
					NegativeOrZero(R[ld])
				case 0x1:
					// MUL
					R[ld] *= R[lm]
					SetCarryAndOverflow(R[ld]*(R[lm]-1), R[lm], "+")
					NegativeOrZero(R[ld]*(R[lm]-1) + R[lm])
				case 0x2:
					//BIC
					R[ld] &^= R[lm]
				case 0x3:
					//MOVN
					R[ld] = 0xFFFFFFFF ^ R[lm]
					NegativeOrZero(R[ld])
				}
			case 0x4:
				Ld := instruction.Value & 0x0007
				Hm := instruction.Value & 0x0038 >> 3
				R[Ld] += R[Hm]
			case 0x5:
				// BIT 7 - 6
				// CMP
				Ld := instruction.Value & 0x0007
				Hm := instruction.Value & 0x0038 >> 3
				SetCarryAndOverflow(R[Ld], R[Hm], "-")
				NegativeOrZero(R[Ld] - R[Hm])
			case 0x6:
				switch (instruction.Value & 0x00C0) >> 6 {
				// BIT 7 - 6
				case 0x0:
					// CPY - Don't update flags
					Ld := instruction.Value & 0x0007
					Hm := instruction.Value & 0x0038 >> 3
					R[Ld] = R[Hm]
				// MOV UPDATE Z E N FLAGS
				default:
					Ld := instruction.Value & 0x0007
					Hm := instruction.Value & 0x0038 >> 3
					R[Ld] = R[Hm]
					NegativeOrZero(R[Hm])
				}
			case 0x7:
				// BIT 7
				if (instruction.Value&0x0080)>>7 == 0 {
				} else {

				}
			}
		} else {
			// LDR
			ld := instruction.Value & 0x0700 >> 8
			immed8 := instruction.Value & 0x00FF
			immed8 *= 4
			R[ld] = int(LDR(PC + immed8))
		}
	case 0x5:
		op := instruction.Value & 0x0600 >> 9
		// 0000 0110 0000 0000
		lm := instruction.Value & 0x01C0 >> 6
		// 0000 0001 1100 0000
		ln := instruction.Value & 0x0038 >> 3
		// 0000 0000 0011 1000
		ld := instruction.Value & 0x0007
		// 0000 0000 0000 0111
		if (instruction.Value&0x0800)>>11 == 0 {
			//STR | STRH | STRB | LDRSB
			switch op {
			case 0:
				//STR
				STR(R[ld], R[lm]+R[ln])
			case 1:
				STRH(R[ld], int16(R[lm]+R[ln]))
				//STRH
			case 2:
				STRB(R[ld], int8(R[lm]+R[ln]))
				//STRB
			case 3:
				R[ld] = int(LDRSB(R[ln] + R[lm]))
				//LDRSB
			}
		} else {
			//LDR | LDRH | LDRB | LDRSH
			switch op {
			case 0:
				//LDR
				R[ld] = int(LDR(R[lm] + R[ln]))
			case 1:
				R[ld] = int(LDRH((R[lm] + R[ln])))
				//LDRH
			case 2:
				R[ld] = int(LDRB((R[lm] + R[ln])))
				//LDRB
			case 3:
				R[ld] = int(LDRSH(R[ln] + R[lm]))
				//LDRSH
			}
		}
	case 0x6:
		ln := instruction.Value & 0x0038 >> 3
		ld := instruction.Value & 0x0007
		immed5 := instruction.Value & 0x07C0 >> 6
		if (instruction.Value&0x0800)>>11 == 0 {
			//STR
			STR(R[ld], R[ln]+immed5*4)
		} else {
			R[ld] = int(LDR(R[ln] + immed5*4))
			//LDR Ld, [Ln, #immed*4]
		}
	case 0x7:
		ln := instruction.Value & 0x0038 >> 3
		ld := instruction.Value & 0x0007
		immed5 := instruction.Value & 0x07C0 >> 6
		if (instruction.Value&0x0800)>>11 == 0 {
			STRB(R[ld], int8(R[ln]+immed5))
			//STRB
		} else {
			R[ld] = int(LDRB(R[ln] + immed5))
			// LDRB Ld, [Ln, #immed]
		}
	case 0x8:
		ln := instruction.Value & 0x0038 >> 3
		ld := instruction.Value & 0x0007
		immed5 := instruction.Value & 0x07C0 >> 6
		if (instruction.Value&0x0800)>>11 == 0 {
			STRH(R[ld], int16(R[ln]+immed5*2))
			//STRH
		} else {
			R[ld] = int(LDRH(R[ln] + immed5*2))
			//LDRH Ld, [Ln, #immed*2]
		}
	case 0x9:
		immed8 := instruction.Value & 0x00FF
		ld := instruction.Value & 0x0380 >> 7
		if (instruction.Value&0x0800)>>11 == 0 {
			//STR
			STR(R[ld], immed8*4+SP)
		} else {
			// LDR Ld, [sp, #immed*4]
			R[ld] = int(LDR(immed8*4 + SP))
		}

	case 0xA:
		// BIT 11
		ld := (instruction.Value & 0x0700) >> 8
		immed8 := instruction.Value & 0x00FF

		if (instruction.Value&0x0800)>>11 == 0 {
			//ADD Ld, pc, #immed*4
			R[ld] = PC + immed8*4
		} else {
			//ADD Ld, sp, #immed*4
			R[ld] = SP + immed8*4
		}
	case 0xB:
		// BIT 11
		if (instruction.Value&0x0800)>>11 == 0 {
			switch (instruction.Value & 0x0600) >> 9 {
			case 0x0:
				immed := 0x007F
				if instruction.Value&0x0080>>7 == 0 {
					//ADD sp, #immed*4
					SP += immed * 4
				} else {
					SP -= immed * 4
					//SUB sp,
					//#immed*4
				}
			case 0x1:
				//SXTH | SXTB | UXTH | UXTB
				op := instruction.Value & 0x00C0 >> 6
				lm := instruction.Value & 0x0038 >> 3
				ld := instruction.Value & 0x0007
				if op > 0 && op < 2 {
					R[ld] = int(int32(R[ld]))
				} else {
					for j := 0; j < R[lm]; j++ {
						if R[ld]&0x00000001 == 1 {
							R[ld] = int(int32(R[ld] >> 1))
							R[ld] |= 0x80000000
						} else {
							R[ld] = int(int32(R[ld] >> 1))
						}
					}
				}

			case 0x2:
				// PUSH
				tmp := instruction.Value & 0x00FF
				for k := 0; k <= 7; k++ {
					if (tmp>>k)&0x1 == 1 {
						Push(R[k])
					}
				}
				// R = 1; IR estiver no register list
			case 0x3:
				//SETEND LE | SETEND BE
				//NÃO PRECISA SER IMPLEMENTADA

			}
		} else {
			switch (instruction.Value & 0x0600) >> 9 {
			case 0x1:
			case 0x2:
				// POP
				tmp := instruction.Value & 0x00FF

				for k := 0; k <= 7; k++ {
					if (tmp>>k)&0x1 == 1 {
						R[k] = Pop()
					}
				}
			case 0x3:
				//BreakPoint
				PrintState()
				fmt.Println("--------- Memória de Dados  -------")
				memory.DataMem()
			}
		}
	case 0xC:
		ln := instruction.Value & 0x0700 >> 8

		if instruction.Value&0x0800>>11 == 0 {
			//STDMIA
			tmp := instruction.Value & 0x00FF
			for k := 0; k <= 7; k++ {
				if (tmp>>k)&0x1 == 1 {
					if STR(R[ln], R[k]) == false {
						fmt.Println("Data Abort")
						break
					}

				}
			}
		} else {
			//LDMIA
			tmp := instruction.Value & 0x00FF
			for k := 0; k <= 7; k++ {
				if (tmp>>k)&0x1 == 1 {
					R[k] = int(LDR(R[ln]))
				}
			}
		}
	case 0xD:
		if instruction.Value&0x0F00>>8 == 0xF {
			if !isUpper(CPSR.Status[6]) {
				fmt.Println("Ocorreu uma interrupção SWI", instruction.Value&0x00FF)
			}
		}
		//B <COND>
		// BIT 11 - 8
		offset := instruction.Value & 0x00FF
		if instruction.Value&0x0080>>7 == 1 {
			offset |= 0x80
			offset ^= 0xFF
			offset += 1
			offset *= -1
		}
		tmp := PC + 2 + 2*offset

		switch instruction.Value & 0x0F00 >> 8 {

		case 0:
			// EQ
			// Z
			if isUpper(CPSR.Status[1]) {
				PC = tmp
			}
		case 1:
			// NE
			// z
			if !isUpper(CPSR.Status[1]) {
				PC = tmp
			}
		case 2:
			// CS/HS
			// C
			if isUpper(CPSR.Status[2]) {
				PC = tmp
			}
		case 3:
			// CC/LO
			// c
			if !isUpper(CPSR.Status[2]) {
				PC = tmp
			}
		case 4:
			//MI
			// N
			if isUpper(CPSR.Status[0]) {
				PC = tmp
			}
		case 5:
			//PL
			// n
			if !isUpper(CPSR.Status[0]) {
				PC = tmp
			}
		case 6:
			//VS
			// V
			if isUpper(CPSR.Status[3]) {
				PC = tmp
			}
		case 7:
			//VC
			// v
			if !isUpper(CPSR.Status[3]) {
				PC = tmp
			}
		case 8:
			//HI
			// z & C
			if !isUpper(CPSR.Status[1]) && isUpper(CPSR.Status[2]) {
				PC = tmp
			}
		case 9:
			//LS
			// Z | c
			if isUpper(CPSR.Status[1]) || !isUpper(CPSR.Status[2]) {
				PC = tmp
			}
		case 0xA:
			//GE
			// NV | nv
			if isUpper(CPSR.Status[0]) && isUpper(CPSR.Status[3]) || !isUpper(CPSR.Status[0]) && !isUpper(CPSR.Status[3]) {
				PC = tmp
			}
		case 0xB:
			//LT
			// Nv | nV
			if isUpper(CPSR.Status[0]) && !isUpper(CPSR.Status[3]) || !isUpper(CPSR.Status[0]) && isUpper(CPSR.Status[3]) {
				PC = tmp
			}
		case 0xC:
			//GT
			// NzV | nzv
			if isUpper(CPSR.Status[0]) && !isUpper(CPSR.Status[1]) && isUpper(CPSR.Status[3]) || !isUpper(CPSR.Status[0]) && !isUpper(CPSR.Status[3]) && !isUpper(CPSR.Status[0]) {
				PC = tmp
			}
		case 0xD:
			//LE
			// Z | Nv | nV
			if isUpper(CPSR.Status[1]) || isUpper(CPSR.Status[0]) && !isUpper(CPSR.Status[3]) || !isUpper(CPSR.Status[0]) && isUpper(CPSR.Status[0]) && isUpper(CPSR.Status[3]) {
				PC = tmp
			}
		}
	case 0xE:
		if instruction.Value&0x0700>>11 == 0 {

			// B
			tmp := instruction.Value & 0x07FF

			if instruction.Value&0x0400>>10 == 1 {
				tmp |= 0x800
				tmp ^= 0xFFF
				tmp += 1
				tmp *= -1
			}

			// PC + 2 pois o PC já incrementou antes da decodificação
			PC = PC + 2 + 2*tmp
		} else {
			// BLX
		}
	case 0xF:
		// BL
		LR = instruction.Address + 2
		tmp := instruction.Value & 0x07FF
		if instruction.Value&0x0400>>10 == 1 {
			tmp |= 0x800
			tmp ^= 0xFFF
			tmp += 1
			tmp *= -1
		}
		// PC + 2 pois o PC já incrementou antes da decodificação
		PC = PC + 2 + 2*tmp
	}
	return true
}
