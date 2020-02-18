// This package implements the quiz, with a variation on part 2.  
// Instead of having an timer for the entire test, it handles a timer 
// for each question.  This was more challenging, since characters 
// entered for previous questions had to be properly discarded
// to get the desired behavior.

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type QuizItem struct {
	Question string
	Answer   string
}

type CharWithTime struct {
	char string
	time time.Time
}

type ErrTimeout struct {
	message string
}

func (e *ErrTimeout) Error() string {
	return e.message
}

func NewErrorTimeout(message string) *ErrTimeout {
	return &ErrTimeout{message: message}
}

func OldReadQuizFromFile(filepath string) []QuizItem {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	items := make([]QuizItem, 0, 100)
	fscanner := bufio.NewScanner(file)
	for fscanner.Scan() {
		line := fscanner.Text()
		// fmt.Println(line)
		tokens := strings.Split(line, ",")
		items = append(items, QuizItem{Question: tokens[0], Answer: tokens[1]})
	}
	return items
}

func ReadQuizFromFile(filepath string) []QuizItem {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	items := make([]QuizItem, 0, 100)
	r := csv.NewReader(file)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(record)
		items = append(items, QuizItem{Question: record[0], Answer: record[1]})
	}
	return items
}

func ReadCharsWithTime(chars chan CharWithTime) {
	reader := bufio.NewReader(os.Stdin)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		chars <- CharWithTime{char: string(r), time: time.Now()}
	}
}

func ReadStringWithTimeout(chars chan CharWithTime, timeout time.Duration) (string, error) {
	start := time.Now()
	// end := start.Add(timeout)
	result := ""
	for {
		select {
		case c := <-chars:
			if c.time.Before(start) {
				// Just ignore old input
				continue
			}
			if c.char == "\n" {
				// Found end of string
				return result, nil
			}
			// Otherwise add character to result buffer.
			result += c.char
		case <-time.After(timeout):
			return "", NewErrorTimeout("after %v seconds")
		}
	}

}

func ReadStringFromStdioWithTimeout(timeout time.Duration) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var result string
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		if string(r) == "\n" {
			break
		}
		result += string(r)
	}
	return result, nil
}

func HandleQuestion(i int, quizItem QuizItem, timeOut time.Duration, chars chan CharWithTime) (correct bool) {
	fmt.Printf("  Question %v %v: ", i, quizItem.Question)

	response, err := ReadStringWithTimeout(chars, timeOut)

	if err != nil {
		switch err.(type) {
		case *ErrTimeout:
			fmt.Println("\nYou didn't answer in time!  The correct answer is " + quizItem.Answer)
			return false
		default:
			log.Fatal(err)
		}
	}
	response = strings.TrimSpace(response)
	if response == quizItem.Answer {
		fmt.Println("Right!")
		return true
	} else {
		fmt.Println("Wrong!  The correct answer is " + quizItem.Answer)
		return false
	}
}

func main() {
	filepath := "problems.csv"
	args := os.Args
	if len(args) > 1 {
		filepath = args[1]
	}
	fmt.Printf("File name: %v\n", filepath)

	timeOutSeconds := 10

	items := ReadQuizFromFile(filepath)
	fmt.Println("Parsed values:")
	for i, value := range items {
		fmt.Printf("  %v %+v \n", i, value)
	}

	nTotal := len(items)
	nCorrect := 0
	chars := make(chan CharWithTime, 100)
	go ReadCharsWithTime(chars)
	fmt.Println("Quiz:")
	for i, value := range items {
		if HandleQuestion(i, value, time.Duration(timeOutSeconds)*time.Second, chars) {
			nCorrect++
		}
	}
	fmt.Printf("You got %v correct out a total of %v questions\n", nCorrect, nTotal)
}
