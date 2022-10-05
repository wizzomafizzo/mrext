package main

import (
	"fmt"

	"github.com/wizzomafizzo/mrext/pkg/input"
)

func main() {
	gps, err := input.GetGamepads()
	if err != nil {
		panic(err)
	}

	for _, gp := range gps {
		fmt.Println("---")
		fmt.Printf("Name:            %s\n", gp.Name)
		fmt.Printf("uinput device:   %s\n", gp.EvDev)
		fmt.Printf("Joystick device: %v\n", gp.JsDev)
		fmt.Printf("Product ID:      %s\n", gp.Product)
		fmt.Printf("Vendor ID:       %s\n", gp.Vendor)
		fmt.Printf("MAC address:     %s\n", gp.Mac)
		fmt.Printf("Battery:         %d%%\n", gp.Battery)
	}
}
