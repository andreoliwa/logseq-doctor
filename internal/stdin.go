package internal

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// ReadFromStdin reads from stdin and returns the content as a string.
// It doesn't return an error and aborts the program if it fails because it's an internal function.
func ReadFromStdin() string {
	scanner := bufio.NewScanner(os.Stdin)

	var stdin string

	var stdinSb16 strings.Builder
	for scanner.Scan() {
		stdinSb16.WriteString(scanner.Text() + "\n")
	}

	stdin += stdinSb16.String()

	err := scanner.Err()
	if err != nil {
		log.Fatalln("error reading input from stdin:", err)
	}

	return stdin
}
