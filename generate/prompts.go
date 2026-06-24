package generate

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PromptForChoice displays a numbered list of options and returns the selected index (0-based)
func PromptForChoice(prompt string, options []string) int {
	fmt.Println(prompt)
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}
	fmt.Print("Enter selection: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(options) {
		fmt.Printf("Invalid selection. Please enter a number between 1 and %d\n", len(options))
		return PromptForChoice(prompt, options)
	}

	return selection - 1
}

// PromptForString prompts for a string input and returns it
func PromptForString(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
