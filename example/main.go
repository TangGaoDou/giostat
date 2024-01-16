package main

import (
	"fmt"

	"github.com/TangGaoDou/giostat"
)

func main() {
	fmt.Println(float64(giostat.ScClkTck))
}
