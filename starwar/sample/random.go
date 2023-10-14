package sample

import (
	"math/rand"

	"github.com/google/uuid"
)

func RandomPrice() float32 {
	return float32(rand.Intn(1000))
}

func RandomId() string {
	return uuid.New().String()
}
