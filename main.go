package main

type Chip8 struct {
	memory     [4096]byte
	display    [32][64]bool
	pc         uint16
	i          uint16
	stack      []uint16
	delayTimer uint8
	soundTimer uint8
	v          [16]byte
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
		c.memory[fontStartAddr + i] = char
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
	opcode := uint16(c.memory[c.pc]) << 8 | uint16(c.memory[c.pc + 1])
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
		}
	case 0x1: // 1NNN - JP addr: jump to NNN
		c.pc = nnn
	case 0x6: // 6XNN - LD Vx, byte: set VX to NN
		c.v[x] = byte(nn)
	case 0x7: // 7XNN - ADD Vx, byte: add NN to VX
		c.v[x] += byte(nn)
	case 0xA: // ANNN - LD I, addr: set I to NNN
		c.i = nnn
	case 0xD: // DXYN - DRW Vx, Vy, nibble: draw N-byte sprite at (VX, VY), set VF on collision
		screenX := uint16(c.v[x] & 63)
		screenY := uint16(c.v[y] & 31)
		c.v[0xF] = 0
		for row := 0; row < int(n); row++ {
			spriteByte := c.memory[c.i + uint16(row)]
			pixelY := screenY + uint16(row)
			if pixelY > 31 {
				break
			}
			for col := range 8 {
				pixelX :=  screenX + uint16(col)
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
	}	
}