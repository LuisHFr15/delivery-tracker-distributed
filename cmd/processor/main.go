package main

import "fmt"

func main() {
	fmt.Println("processor starting...")
	select {} // Keep running
}

