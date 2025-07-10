package internal

import (
	"bufio"
	"log"
	"os"
)

// ReadFromStdin reads from stdin and returns the content as a string.
// It doesn't return an error and aborts the program if it fails because it's an internal function.
func ReadFromStdin() string {
	scanner := bufio.NewScanner(os.Stdin)

	var stdin string

	for scanner.Scan() {
		stdin += scanner.Text() + "\n"
	}

	err := scanner.Err()
	if err != nil {
		log.Fatalln("error reading input from stdin:", err)
	}

	return stdin
}
