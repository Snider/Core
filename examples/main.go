package main

import (
	"fmt"

	"github.com/Snider/Core"
)

func main() {
	// Display library version
	fmt.Printf("Core Library Version: %s\n\n", core.GetVersion())

	// Example 1: Using the Greeter
	fmt.Println("Example 1: Greeter")
	greeter := core.NewGreeter("World")
	fmt.Println(greeter.Greet())

	greeter2 := core.NewGreeter("Go Developer")
	fmt.Println(greeter2.Greet())
	fmt.Println()

	// Example 2: Using the Add function
	fmt.Println("Example 2: Add function")
	sum := core.Add(10, 20)
	fmt.Printf("10 + 20 = %d\n", sum)
	fmt.Println()

	// Example 3: Using the Multiply function
	fmt.Println("Example 3: Multiply function")
	product := core.Multiply(7, 8)
	fmt.Printf("7 * 8 = %d\n", product)
	fmt.Println()

	// Example 4: Chaining operations
	fmt.Println("Example 4: Chaining operations")
	a := core.Add(5, 3)
	b := core.Multiply(a, 2)
	fmt.Printf("(5 + 3) * 2 = %d\n", b)
}
