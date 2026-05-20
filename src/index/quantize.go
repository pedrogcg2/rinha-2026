package index

const Scale = 10_000

func Quantize(x float64) int16 {
	d := x * Scale
	if d < -Scale {
		return -Scale
	}
	if d > Scale {
		return Scale
	}
	return int16(d)
}
