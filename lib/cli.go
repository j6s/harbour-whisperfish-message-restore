package lib

import (
	"bufio"
	"log"
	"os"
)

func AskConfirm(question string) bool {
	log.Printf("%s [y|N]", question)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "y" || text == "Y" {
			return true
		}
		if text == "n" || text == "N" {
			return false
		}

		log.Print(question)
	}

	return false
}
