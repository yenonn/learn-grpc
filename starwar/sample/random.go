package sample

import "math/rand"

func RandomPrice() float32 {
	return float32(rand.Intn(1000))
}
