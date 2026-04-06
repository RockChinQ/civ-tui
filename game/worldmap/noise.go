package worldmap

func hash2(x, y int, seed int64) float64 {
	h := int64(x)*1000003 ^ int64(y)*999983 ^ seed
	h = h ^ (h >> 13)
	h *= -1640531527
	h = h ^ (h >> 15)
	if h < 0 {
		h = -h
	}
	return float64(h%10000) / 10000.0
}

func smoothNoise(x, y float64, seed int64) float64 {
	ix, iy := int(x), int(y)
	fx := x - float64(ix)
	fy := y - float64(iy)
	fx = fx * fx * (3 - 2*fx)
	fy = fy * fy * (3 - 2*fy)
	v00 := hash2(ix, iy, seed)
	v10 := hash2(ix+1, iy, seed)
	v01 := hash2(ix, iy+1, seed)
	v11 := hash2(ix+1, iy+1, seed)
	return v00*(1-fx)*(1-fy) + v10*fx*(1-fy) + v01*(1-fx)*fy + v11*fx*fy
}

func fractalNoise(x, y int, seed int64) float64 {
	total, maxVal := 0.0, 0.0
	amp, freq := 1.0, 1.0
	scale := 0.05
	for i := 0; i < 4; i++ {
		total += smoothNoise(float64(x)*scale*freq, float64(y)*scale*freq, seed+int64(i)*1234567) * amp
		maxVal += amp
		amp *= 0.5
		freq *= 2.0
	}
	return total / maxVal
}
