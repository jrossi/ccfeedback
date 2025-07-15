package ccfeedback

import "fmt"

func TestVisibility() {
	// This should trigger warnings
	var x string = "forbidden interface{}"
	fmt.Println(x)

	// Bad formatting
	fmt.Println("bad indent")
}
