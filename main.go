package main

import "math/rand/v2"

type Chip8 struct {
	memory     [4096]byte
	display    [32][64]bool
	pc         uint16
	i          uint16
	stack      []uint16
	delayTimer uint8
	soundTimer uint8
	v          [16]byte
	keypad 	   [16]bool
}

var font = [80]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

const fontStartAddr = 0x050

func main() {
	c := NewChip8()

	// Load font to memory 050-09F (popular convention)
	for i, char := range font {
		c.memory[fontStartAddr+i] = char
	}
}

func NewChip8() *Chip8 {
	const romStartAddr = 0x200
	c := Chip8{
		pc: romStartAddr,
	}
	return &c
}

func (c *Chip8) Cycle() {
	// Shift 8 bits and do OR operation to combine to a full instruction
	opcode := uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])
	c.pc += 2

	op := (opcode & 0xF000) >> 12
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	n := opcode & 0x000F
	nn := opcode & 0x00FF
	nnn := opcode & 0x0FFF
	switch op {
	case 0x0:
		switch nnn {
		case 0x0E0: // 00E0 - CLS: clear display
			c.display = [32][64]bool{}
		case 0x0EE: // 00EE - RET: return from subroutine
			c.pc = c.stack[len(c.stack)-1]
			c.stack = c.stack[:len(c.stack)-1]
		}
	case 0x1: // 1NNN - JP addr: jump to NNN
		c.pc = nnn
	case 0x2: // 2NNN - CALL addr: call subroutine at nnn
		c.stack = append(c.stack, c.pc)
		c.pc = nnn
	case 0x3: // 3XNN - SE Vx, byte: skip next instruction if VX == NN
		if c.v[x] == byte(nn) {
			c.pc += 2
		}
	case 0x4: // 4XNN - SNE Vx, byte: skip next instruction if VX != NN
		if c.v[x] != byte(nn) {
			c.pc += 2
		}
	case 0x5: // 5XY0 - SE Vx, byte: skip next instruction if VX == VY
		if c.v[x] == c.v[y] {
			c.pc += 2
		}
	case 0x6: // 6XNN - LD Vx, byte: set VX to NN
		c.v[x] = byte(nn)
	case 0x7: // 7XNN - ADD Vx, byte: add NN to VX
		c.v[x] += byte(nn)
	case 0x8:
		switch n {
		case 0x0: // 8XY0 - LD Vx, Vy: set VX to VY
			c.v[x] = c.v[y]
		case 0x1: // 8XY1 - OR Vx, Vy: set VX to VX OR VY
			c.v[x] = c.v[x] | c.v[y]
		case 0x2: // 8XY2 - AND Vx, Vy: set VX to VX AND VY
			c.v[x] = c.v[x] & c.v[y]
		case 0x3: // 8XY3 - XOR Vx, Vy: set VX to VX XOR VY
			c.v[x] = c.v[x] ^ c.v[y]
		case 0x4: // 8XY4 - ADD Vx, Vy: VX += VY, VF set to 1 on overflow, else 0
			sum := uint16(c.v[x]) + uint16(c.v[y])
			if sum > 0xFF {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] = byte(sum)
		case 0x5: // 8XY5 - SUB Vx, Vy: VX -= VY, VF set to 1 if no borrow (VX >= VY), else 0
			if c.v[x] >= c.v[y] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] = c.v[x] - c.v[y]
		case 0x6: // 8XY6 - SHR Vx: shift VX right 1 bit, VF set to bit shifted out
			c.v[x] = c.v[y] // Original chip8 intrepreter behaviour for COSMAC VIP
			shiftedBit := c.v[x] & 0x1
			c.v[x] = c.v[x] >> 1
			c.v[0xF] = shiftedBit
		case 0x7: // 8XY7 - SUB Vx, Vy: VX = VY - VX, VF set to 1 if no borrow (VY >= VX), else 0
			if c.v[y] >= c.v[x] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] = c.v[y] - c.v[x]
		case 0xE: // 8XYE - SHL Vx: shift VX left 1 bit, VF set to bit shifted out
			c.v[x] = c.v[y] // Original chip8 intrepreter behaviour for COSMAC VIP
			shiftedBit := (c.v[x] & 0x80) >> 7
			c.v[x] = c.v[x] << 1
			c.v[0xF] = shiftedBit
		}
	case 0x9: // 9XY0 - SNE Vx, byte: skip next instruction if VX != VY
		if c.v[x] != c.v[y] {
			c.pc += 2
		}
	case 0xA: // ANNN - LD I, addr: set I to NNN
		c.i = nnn
	case 0xB: // BNNN - JP V0, addr: jump to NNN + V0 (original COSMAC VIP behaviour)
		c.pc = nnn + uint16(c.v[0])
	case 0xC: // CXNN - RND Vx, byte: set VX to a random byte ANDed with NN
		randVal := rand.IntN(256)
		c.v[x] = byte(randVal) & byte(nn)
	case 0xD: // DXYN - DRW Vx, Vy, nibble: draw N-byte sprite at (VX, VY), set VF on collision
		screenX := uint16(c.v[x] & 63)
		screenY := uint16(c.v[y] & 31)
		c.v[0xF] = 0
		for row := 0; row < int(n); row++ {
			spriteByte := c.memory[c.i+uint16(row)]
			pixelY := screenY + uint16(row)
			if pixelY > 31 {
				break
			}
			for col := range 8 {
				pixelX := screenX + uint16(col)
				spriteBit := spriteByte & (0x80 >> col)
				if pixelX > 63 {
					break
				}
				if spriteBit != 0 && c.display[pixelY][pixelX] == true {
					c.display[pixelY][pixelX] = false
					c.v[0xF] = 1
				} else if spriteBit != 0 && c.display[pixelY][pixelX] == false {
					c.display[pixelY][pixelX] = true
				}

			}
		}
	case 0xE:
		switch nn {
		case 0x9E: // EX9E - SKP Vx: skip next instruction if key VX is pressed
			if c.keypad[c.v[x]] {
				c.pc += 2
			}
		case 0xA1: // EXA1 - SKP Vx: skip next instruction if key VX is not pressed
			if !c.keypad[c.v[x]] {
				c.pc += 2
			}
		}
	case 0xF:
		switch nn {
		case 0x07: // FX07 - LD Vx, DT: set VX to the value of the delay timer
			c.v[x] = c.delayTimer
		case 0x15: // FX15 - LD DT, Vx: set the delay timer to VX
			c.delayTimer = c.v[x]
		case 0x18: // FX18 - LD ST, Vx: set the sound timer to VX
			c.soundTimer = c.v[x]
		case 0x1E: // FX1E - ADD I, Vx: add VX to I
			c.i += uint16(c.v[x])
		case 0x0A: // FX0A - LD Vx, K: wait for a key press, store the key's value in VX (blocking)
			keyPressed := false
			for i, key := range c.keypad {
				if key {
					c.v[x] = byte(i)
					keyPressed = true
					break
				}
			}
			if !keyPressed {
				c.pc -= 2
			}
		case 0x29: // FX29 - LD F, Vx: set I to the memory address of the font character in VX
			lastNibble := c.v[x] & 0xF
			c.i = uint16(fontStartAddr + (lastNibble * 5))
		case 0x33: // FX33 - LD B, Vx: store BCD representation of VX in memoery I, I + 1, I + 2
			decimal := int(c.v[x])
			ones := decimal % 10
			decimal = decimal / 10
			tens := decimal % 10
			decimal = decimal / 10
			hundreds := decimal % 10
			c.memory[c.i] = byte(hundreds)
			c.memory[c.i + 1] = byte(tens)
			c.memory[c.i + 2] = byte(ones)
		case 0x55: // FX55 - LD [I], Vx: store registers V0 - VX in memory starting at I
			for offset := range x + 1 {
				c.memory[c.i + uint16(offset)] = c.v[offset]
			}
		case 0x65: // FX65 - LD Vx, [I]: load register V0 - VX from memory starting at I
			for offset := range x + 1 {
				c.v[offset] = c.memory[c.i + uint16(offset)]
			}
		}
	}
}
