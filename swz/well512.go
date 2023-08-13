package swz

type well512 struct {
	index int32
	state [16]uint32
}

func newWell512(seed uint32) *well512 {
	w := new(well512)
	w.index = 0
	w.setSeed(seed)
	return w
}

func (w *well512) setSeed(seed uint32) {
	w.index = 0

	w.state[0] = seed

	for i := uint32(1); i < 0x10; i++ {
		w.state[i] = i + 0x6C078965*(w.state[i-1]^(w.state[i-1]>>30))
	}
}

func (w *well512) nextUint() uint32 {
	var a, b, c, d uint32

	a = w.state[w.index]
	c = w.state[(w.index+13)&15]
	b = a ^ c ^ (a << 16) ^ (c << 15)
	c = w.state[(w.index+9)&15]
	c ^= c >> 11
	w.state[w.index] = b ^ c
	a = w.state[w.index]
	d = a ^ ((a << 5) & 0xDA442D24)
	w.index = (w.index + 15) & 15
	a = w.state[w.index]
	w.state[w.index] = a ^ b ^ d ^ (a << 2) ^ (b << 18) ^ (c << 28)

	return w.state[w.index]
}
