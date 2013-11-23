package nes

type Word uint8

type Memory []Word

type MemoryError struct {
	ErrorText string
}

func (e MemoryError) Error() string {
	return e.ErrorText
}

func fitAddressSize(addr interface{}) (v int, e error) {
	switch a := addr.(type) {
	case Word:
		v = int(a)
	case int:
		v = int(a)
	case uint16:
		v = int(a)
	default:
		e = MemoryError{ErrorText: "Invalid type used"}
	}

	return
}

func NewMemory() Memory {
	return make([]Word, 0x10000)
}

func (m Memory) ReadMirroredRam(a int) Word {
	offset := a % 0x8
	return m[0x2000+offset]
}

func (m Memory) WriteMirroredRam(v Word, a int) {
	offset := a % 0x8
	m[0x2000+offset] = v
}

func (m Memory) Write(address interface{}, val Word) error {
	if a, err := fitAddressSize(address); err == nil {
		if a >= 0x2000 && a <= 0x2007 {
			ppu.RegWrite(val, a)
			// m.WriteMirroredRam(val, a)
		} else if a == 0x4014 {
			ppu.RegWrite(val, a)
			m[a] = val
		} else if a == 0x4016 {
			Pads[0].Write(val)
			m[a] = val
		} else if a == 0x4017 {
			Pads[1].Write(val)
			apu.RegWrite(val, a)
			m[a] = val
		} else if a&0xF000 == 0x4000 {
			apu.RegWrite(val, a)
		} else if a >= 0x8000 && a <= 0xFFFF {
			rom.Write(val, a)
			return nil
		} else if a >= 0x5100 && a <= 0x6000 {
			if v, ok := rom.(*Mmc5); ok {
				// MMC5 register handling
				v.Write(val, a)
				return nil
			}

			m[a] = val
		} else {
			m[a] = val
		}

		return nil
	}

	return MemoryError{ErrorText: "Invalid address used"}
}

func (m Memory) Read(a uint16) (Word, error) {
	switch {
	case a >= 0x2008 && a < 0x4000:
		offset := a % 0x8
		return ppu.RegRead(int(0x2000 + offset))
	case a <= 0x2007 && a >= 0x2000:
		return ppu.RegRead(int(a))
	case a == 0x4016:
		return Pads[0].Read(), nil
	case a == 0x4017:
		return Pads[1].Read(), nil
	case a&0xF000 == 0x4000:
		return apu.RegRead(int(a))
	case a >= 0x8000 && a <= 0xFFFF:
		return rom.Read(int(a)), nil
	case a >= 0x5100 && a <= 0x6000:
		if _, ok := rom.(*Mmc5); ok {
			// MMC5 register handling
			return rom.Read(int(a)), nil
		}
	}

	return m[a], nil
}
