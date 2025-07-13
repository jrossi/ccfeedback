package ccfeedback

import "fmt"

// ProperFunction demonstrates proper Go formatting
func ProperFunction() {
	message := "This is properly formatted"
	fmt.Println(message)

	count := 10
	if count > 5 {
		fmt.Printf("Count is %d\n", count)
	}
}

// ProperHelper returns a greeting
func ProperHelper(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
