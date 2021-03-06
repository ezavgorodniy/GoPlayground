package main

import (
	"fmt"
	"log"

	"example.com/greetings"
)

func main() {
	// Set properties of the predefined Logger, including
	// the log entry prefix and a flag to disable printing
	// the time, source file, and line number.
	log.SetPrefix("greetings: ")
	log.SetFlags(0)

	// A slice of names.
	anotherNames := []string{"Gladys", "Samantha", "Darrin"}
	anotherNames2 := []string{"Gladys", "Samantha", "Darrin"}
	anotherNames3 := []string{"Gladys", "Samantha", "Darrin"}

	// Request greeting messages for the names.
	receivedMessages, err := greetings.Hellos(anotherNames)
	if err != nil {
		log.Fatal(err)
	}
	// If no error was returned, print the returned map of
	// messages to the console.
	fmt.Println(receivedMessages)
	fmt.Println(anotherNames2)
	fmt.Println(anotherNames3)
}
