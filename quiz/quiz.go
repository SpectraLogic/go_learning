/*
	quiz defines a quiz and its questions and provides methods for starting and processing quizzes as well as reading in
    a Quiz from file.
*/
package quiz

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

// A Quiz struct is a collection of up to 100 Questions, the name of the file containing the quiz, the number of
// questions in the quiz and the number of questions the user got correct.
type Quiz struct {
	Questions        []Question
	Filename         string
	TotalQuestions   int
	CorrectQuestions int
}


// NewQuiz creates a new Quiz struct based on the information in the file at the given file name
func NewQuiz(filename string) (*Quiz, error) {
	fmt.Println("\nGetting quiz using file:\t", filename)

	// open the file and save Question to a Quiz struct
	csvFile, err := os.Open(filename)

	if err != nil {
		fmt.Println("Error opening CSV file:\t", err)
		return nil, err
	}

	defer func(){
		if csvErr := csvFile.Close(); csvErr != nil {
			fmt.Println(csvErr)
		}
	}()

	reader := csv.NewReader(bufio.NewReader(csvFile))

	questions := make([]Question, 100)
	// loop until end of file or error is reached
	for {
		record, ioErr := reader.Read()
		if ioErr == io.EOF {
			break
		}
		if ioErr != nil {
			fmt.Println("Error reading from file:\t", ioErr)
			return nil, ioErr
		}

		questions = append(questions, Question{
			QuestionText: record[0],
			Answer:       record[1],
			UserAnswer:   "empty",
			Correct:      nil,
		})
	}

	q := Quiz{
		Questions:        questions,
		Filename:         filename,
		TotalQuestions:   len(questions),
		CorrectQuestions: 0,
	}

	return &q, err
}

// Print prints a quiz to stdout with nicely aligned columns thanks to the tabWriter package. Useful for debugging.
func (q Quiz) Print() (err error) {
	fmt.Println("\nQuiz - ", q.Filename)
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
	_, err = fmt.Fprintln(w, "Question\tAnswer\tUser Answer\tCorrect")
	if err != nil {
		return err
	}

	for index, value := range q.Questions {
		quest := q.Questions[index]
		if quest.QuestionText != "" {
			_, err = fmt.Fprintln(w, value)
			if err != nil {
				return err
			}
		}
	}

	_, err = fmt.Fprintln(w)
	if err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}

// Start begins the user-input portion of the quiz. Each Question is presented to the user and they enter answers via
// stdin with "\n" being used to specify the end of an answer.
func (q *Quiz) Start(maxQuizTime int) error {
	var err error

	c := make(chan string)
	successString := takeAndTimeQuiz(q, maxQuizTime, c)
	if successString == "success" {
		return err
	} else {
		return errors.New("quiz timed out before user completed it")
	}
}

// takeAndTimeQuiz proctors the quiz, asking the user questions and saving the responses.
// A Quiz pointer is used as input and output so we can be sure we're looking at the same quiz and not a copy.
// timer is the time in seconds to allow the quiz to run for. If the user doesn't finish in time, the rest of the answers
// are considered incorrect.
func takeAndTimeQuiz(q *Quiz, maxQuizTime int, c chan string) string {
	fmt.Printf("\nYou will have %v seconds to answer %v questions.\nPress [Enter] to Start the quiz.\n",
		maxQuizTime, q.TotalQuestions)
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	// We'll return the first thing down this channel, whether it be the maxQuizTime result or the quiz result. Either way, a
	// pointer to the quiz will be returned so that partial credit can be given.
	ch := make(chan string)
	ch = c

	// start the maxQuizTime
	fmt.Printf("Starting maxQuizTime for %v seconds.\n", maxQuizTime)
	quizTimer := time.NewTimer(time.Duration(maxQuizTime) * time.Second)
	defer quizTimer.Stop()

	// this goroutine waits on the maxQuizTime and sends a "timeout" string down the channel when it is expired.
	// takeAndTimeQuiz returns the first thing down the channel, so the maxQuizTime will 'interrupt' the quiz if it's still going
	go func() {
		<-quizTimer.C
		fmt.Println("\nTimer has run out!")
		ch <- "timeout"
	}()

	// this goroutine performs the quiz, asking the user questions and gathering responses. When it finishes, it sends
	// a "success" message down the channel to be returned to the calling function
	go func() {
		for index, value := range q.Questions {
			if value.QuestionText != "" {
				fmt.Print("\n\tQuestion ", index + 1, ":")
				fmt.Print("\t\t" + value.QuestionText + " = ")

				savedAnswer, _ := reader.ReadString('\n')
				savedAnswer = strings.Replace(savedAnswer, "\n", "", 1)
				result := savedAnswer == value.Answer
				if result {
					fmt.Println("\tCorrect!")
					q.CorrectQuestions++
				} else {
					fmt.Println("\tIncorrect.")
				}
				value.Correct = &result
				value.UserAnswer = savedAnswer
				q.Questions[index] = value
			}
		}
		ch <- "success"
	}()

	return <-ch
}


// Results takes a Quiz and performs the processing needed to determine the % of correct answers
func (q Quiz) Results() string {
	correct, total := float64(q.CorrectQuestions), float64(q.TotalQuestions)
	ratio := 100 * (correct / total)
	percent := "%"
	return fmt.Sprintf("%2.2f%v (%v/%v) of the answers were correct.", ratio, percent, correct, total)
}
