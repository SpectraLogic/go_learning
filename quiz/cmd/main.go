package main

import (
	"flag"
	"fmt"
	"github.com/JayDonigian/go_learning/quiz"
	"os"
	"time"
)

// Read in the quiz from file, start it, wait for user or timer and then print results.
func main() {
	var filename = flag.String("file", "problems.csv", "The file containing the quiz.")
	var random = flag.Bool("random", false,
		"If random is used, the quiz questions will be asked in random order.")
	var timerDuration = flag.Int("time", 30,
		"The amount of time in seconds that the user has to complete the quiz.")
	flag.Parse()

	newQuiz, err := quiz.NewQuiz(*filename)
	errorHandler(err)

	duration := time.Duration(*timerDuration) * time.Second
	err = newQuiz.Start(duration, *random)
	errorHandler(err)

	fmt.Println("\n" + newQuiz.Results())
}

func errorHandler(err error) {
	if err != nil {
		fmt.Printf("An error was encountered: %v", err)
		os.Exit(1)
	}
}
