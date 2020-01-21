package main

import (
	"fmt"
	"github.com/JayDonigian/go_learning/quiz"
	"os"
)

// Read in the quiz from file, start it, wait for user or timer and then print results.
func main() {
	newQuiz, err := quiz.NewQuiz("problems.csv")
	errorHandler(err)

	err = newQuiz.Start(30)
	errorHandler(err)

	fmt.Println("\n" + newQuiz.Results())
}

func errorHandler(err error) {
	if err != nil {
		fmt.Printf("An error was encountered: %v", err)
		os.Exit(1)
	}
}
