package main

import (
	"fmt"

	"github.com/caarlos0/fastcom-exporter/fast"
)

func main() {
	bps, err := fast.Measure()
	fmt.Println(bps/125000, err)
}
