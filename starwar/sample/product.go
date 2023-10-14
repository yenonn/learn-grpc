package sample

import (
	v1 "github.com/yenonn/starwar/pb/v1"
)

func NewProduct() *v1.Product {
	return &v1.Product{
		Id:          "",
		Name:        "",
		Description: "",
		Price:       RandomPrice(),
	}
}
