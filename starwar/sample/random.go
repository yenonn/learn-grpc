package sample

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

func RandomPrice() float32 {
	return float32(rand.Intn(1000))
}

func RandomId() string {
	return uuid.New().String()
}

func RandomNameDescription() (string, string) {
	objectMap := make(map[string]string)
	objectSlice := []string{"book", "pen", "pencil", "eraser", "notebook", "computer"}
	for _, value := range objectSlice {
		objectMap[value] = fmt.Sprint("this is a ", value)
	}
	randomIndex := rand.Intn(len(objectMap))
	return objectSlice[randomIndex], objectMap[objectSlice[randomIndex]]
}
