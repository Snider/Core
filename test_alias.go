package main
import "fmt"
func ErrorFunc(s string) { fmt.Println(s) }
type ErrorAlias = ErrorFunc
func main() {}
