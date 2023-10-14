package sample

import (
	v1 "github.com/yenonn/starwar/pb/v1"
)

func NewProduct() *v1.Product {
	name, description := RandomNameDescription()
	return &v1.Product{
		Id:          RandomId(),
		Name:        name,
		Description: description,
		Price:       RandomPrice(),
	}
}
