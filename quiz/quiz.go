/*
	Quiz defines a quiz and its questions and provides methods for starting and processing quizzes as well as reading in
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

/*
	This Quiz struct is a collection of up to 100 Questions, the name of the file containing the quiz, the number of
questions in the quiz and the number of questions the user got correct.
*/
type Quiz struct {
	Questions        [100]Question
	Filename         string
	TotalQuestions   int
	CorrectQuestions int
}

/*
	Creates a new Quiz struct based on the information in the file at the given file name
*/
func NewQuiz(filename string) (quiz Quiz, err error) {
	fmt.Println("\nGetting quiz using file:\t", filename)

	// open the file and save question to a Quiz struct
	var questionCount = 0
	csvFile, _ := os.Open(filename)
	reader := csv.NewReader(bufio.NewReader(csvFile))

	// loop until end of file or error is reached
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading from file:\t", err)
			return Quiz{}, err
		}

		quiz.Questions[questionCount] = NewQuestion(record[0], record[1])
		questionCount++
	}

	quiz.TotalQuestions = questionCount
	quiz.CorrectQuestions = 0
	quiz.Filename = filename

	return quiz, nil
}

/*
	Print prints a quiz to stdout with nicely aligned columns thanks to the tabwriter package. Useful for debugging.
*/
func (q Quiz) Print() (err error) {
	fmt.Println("\nQuiz - ", q.Filename)
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
	_, err = fmt.Fprintln(w, "Question\tAnswer\tUser Answer\tCorrect")
	if err != nil {
		return err
	}

	for index, value := range q.Questions {
		question := q.Questions[index]
		if question.QuestionText != "" {
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

/*
	Start begins the user-input portion of the quiz. Each question is presented to the user and they enter answers via
stdin with "\n" being used to specify the end of an answer.
*/
func (q Quiz) Start(timer int) (Quiz, error) {
	var err error

	c := make(chan string)
	successString, returnedQuiz := takeAndTimeQuiz(&q, timer, c)
	if successString == "success" {
		return *returnedQuiz, err
	} else {
		return *returnedQuiz, errors.New("quiz timed out before user completed it")
	}

}

/*
	takeAndTimeQuiz proctors the quiz, asking the user questions and saving the responses.
	A Quiz pointer is used as input and output so we can be sure we're looking at the same quiz and not a copy.
    timer is the time in seconds to allow the quiz to run for. If the user doesn't finish in time, the rest of the answers
    are considered incorrect.
*/
func takeAndTimeQuiz(q *Quiz, timer int, c chan string) (string, *Quiz) {
	fmt.Printf("\nYou will have %v seconds to answer %v questions.\nPress [Enter] to Start the quiz.\n",
		timer, q.TotalQuestions)
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	// We'll return the first thing down this channel, whether it be the timer result or the quiz result. Either way, a
	// pointer to the quiz will be returned so that partial credit can be given.
	ch := make(chan string)
	ch = c

	// start the timer
	fmt.Printf("Starting timer for %v seconds.\n", timer)
	quizTimer := time.NewTimer(time.Duration(timer) * time.Second)
	defer quizTimer.Stop()

	// this goroutine waits on the timer and sends a "timeout" string down the channel when it is expired.
	// takeAndTimeQuiz returns the first thing down the channel, so the timer will 'interrupt' the quiz if it's still going
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
				fmt.Print("\n\tQuestion ", index+1, ":")
				fmt.Print("\t\t" + value.QuestionText + " = ")

				reader := bufio.NewReader(os.Stdin)
				storageString, _ := reader.ReadString('\n')
				storageString = strings.Replace(storageString, "\n", "", 1)
				result := storageString == value.Answer
				if result {
					fmt.Println("\tCorrect!")
					q.CorrectQuestions++
				} else {
					fmt.Println("\tIncorrect.")
				}
				value.Correct = &result
				value.UserAnswer = storageString
				q.Questions[index] = value
			}
		}
		ch <- "success"
	}()

	return <-ch, q
}

/*
	Results takes a Quiz and performs the processing needed to determine the % of correct answers
*/
func (q Quiz) Results() string {
	correct, total := float64(q.CorrectQuestions), float64(q.TotalQuestions)
	ratio := 100 * (correct / total)
	percent := "%"
	return fmt.Sprintf("%2.2f%v (%v/%v) of the answers were correct.", ratio, percent, correct, total)
}
