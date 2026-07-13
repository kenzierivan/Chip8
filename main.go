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


