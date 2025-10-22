package main

import (
	"fmt"

	"github.com/Snider/Core"
)

func main() {
	// Create the Core service
	coreService := core.NewCoreService()

	// Display service information
	fmt.Printf("Service: %s\n", coreService.ServiceName())
	fmt.Printf("Version: %s\n\n", coreService.GetVersion())

	// Demonstrate service methods
	fmt.Println("=== Core Service Demo ===\n")

	// Test Greet
	fmt.Println("1. Greeting:")
	fmt.Printf("   Greet('Wails Developer'): %s\n", coreService.Greet("Wails Developer"))
	fmt.Printf("   Greet('Go Programmer'): %s\n\n", coreService.Greet("Go Programmer"))

	// Test Add
	fmt.Println("2. Addition:")
	fmt.Printf("   Add(15, 27): %d\n", coreService.Add(15, 27))
	fmt.Printf("   Add(100, 200): %d\n\n", coreService.Add(100, 200))

	// Test Multiply
	fmt.Println("3. Multiplication:")
	fmt.Printf("   Multiply(8, 9): %d\n", coreService.Multiply(8, 9))
	fmt.Printf("   Multiply(12, 12): %d\n\n", coreService.Multiply(12, 12))

	// Test Calculate with various operations
	fmt.Println("4. Calculator:")
	operations := []struct {
		op string
		a  int
		b  int
	}{
		{"add", 10, 5},
		{"subtract", 20, 8},
		{"multiply", 7, 6},
		{"divide", 100, 4},
	}

	for _, test := range operations {
		result, err := coreService.Calculate(test.op, test.a, test.b)
		if err != nil {
			fmt.Printf("   Calculate('%s', %d, %d): Error - %v\n", test.op, test.a, test.b, err)
		} else {
			fmt.Printf("   Calculate('%s', %d, %d): %d\n", test.op, test.a, test.b, result)
		}
	}

	// Test error handling
	fmt.Println("\n5. Error Handling:")
	_, err := coreService.Calculate("divide", 10, 0)
	if err != nil {
		fmt.Printf("   Calculate('divide', 10, 0): Error - %v ✓\n", err)
	}

	_, err = coreService.Calculate("modulo", 10, 3)
	if err != nil {
		fmt.Printf("   Calculate('modulo', 10, 3): Error - %v ✓\n", err)
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nThis service can be integrated into any Wails v3 application!")
	fmt.Println("For a full UI example, see examples/main.go (requires Wails v3 with UI support)")
}
