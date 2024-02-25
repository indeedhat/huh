package main

import "github.com/charmbracelet/huh"

// TODO: ensure input is a valid fruit
func checkFruitName(s string) error { return nil }

func main() {
	var fruits []string

	text := huh.NewArrayList().
		Title("Name some fruits").
		Validate(checkFruitName).
		Placeholder("What's on your mind?").
		Value(&fruits)

	text.Focus()

	// Create a form to show help.
	form := huh.NewForm(huh.NewGroup(text))
	form.Run()
}
