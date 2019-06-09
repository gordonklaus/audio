package audio

func Crossfade(a, x, b float64) float64 {
	return (1-x)*a + x*b
}
