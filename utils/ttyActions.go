package utils

import (
	"fmt"
	"log"
	"strings"

	"github.com/mattn/go-tty"
)

func AskAction(def bool) bool {
	answer := "?"
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()
	for answer != "y" && answer != "n" && answer != "\n" {
		r, err := tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		answer = string(r)
		answer = strings.ToLower(answer)
	}
	fmt.Printf("%s\n", answer)
	if answer == "y" || (answer == "" && def == true) {
		return true
	}
	return false
}

func AskActionSides(def bool) (bool, string) {
	answer := "?"
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()
	for answer != "<" && answer != ">" && answer != "n" {
		r, err := tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		answer = string(r)
		answer = strings.ToLower(answer)
	}
	fmt.Printf("%s\n", answer)
	if answer == ">" || answer == "<" {
		return true, answer
	}
	return false, ""
}


