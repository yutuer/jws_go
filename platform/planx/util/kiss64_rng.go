package util

//
// The KISS PRNG
//

type Kiss64Rng struct {
	D         [5]uint64
	Count     int
	RandValue int
}

func (k *Kiss64Rng) Seed(seed int64) {
	k.D[1] = (uint64)(seed) & 0x3FFFFFFFFFFFFFF
	k.D[0] = 1234567890987654321
	k.D[2] = 362436362436362436
	k.D[3] = 1066149217761810
	k.D[4] = 0
	k.Count = 1
	k.RandValue = 0

}

func (k *Kiss64Rng) Int63() int64 {
	// MWD[1]
	k.D[4] = (k.D[0] << 58) + k.D[1]
	k.D[1] = (k.D[0] >> 6)
	k.D[0] += k.D[4]
	if k.D[0] < k.D[4] {
		k.D[1]++
	}
	// D[0]SH
	k.D[2] ^= k.D[2] << 13
	k.D[2] ^= k.D[2] >> 17
	k.D[2] ^= k.D[2] << 43
	// D[1]NG
	k.D[3] = 6906969069*k.D[3] + 1234567
	// Result
	res := (int64)((k.D[0] + k.D[2] + k.D[3]) & 0x7FFFFFFFFFFFFFFF)

	k.Count++
	k.RandValue = int(res)

	return res
}
