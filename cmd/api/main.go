package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := loadConfiguration(); err != nil {
		fmt.Println(err)
		return
	}

	r := gin.Default()
	registerEndPoints(r)
	err := r.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
}
