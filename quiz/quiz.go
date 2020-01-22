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
	"math/rand"
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

	defer func() {
		if csvErr := csvFile.Close(); csvErr != nil {
			fmt.Println(csvErr)
		}
	}()

	reader := csv.NewReader(bufio.NewReader(csvFile))

	questions := make([]Question, 0, 100)
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
func (q *Quiz) Start(maxQuizTime time.Duration, random bool) error {
	fmt.Printf("\nYou will have %v seconds to answer %v questions.\nPress [Enter] to Start the quiz.\n",
		maxQuizTime, q.TotalQuestions)

	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	c := make(chan string)
	go timeQuiz(maxQuizTime, c)
	if random {
		go proctorRandomQuiz(q, reader, c)
	} else {
		go proctorQuiz(q, reader, c)
	}

	if <-c == "success" {
		return nil
	} else {
		errorText := fmt.Sprintf("The quiz timer (%v seconds) ran out before the user completed the quiz",
			maxQuizTime)
		return errors.New(errorText)
	}
}

// timeQuiz waits on the maxQuizTime and sends a "timeout" string down the channel when it is expired.
// takeAndTimeQuiz returns the first thing down the channel, so the maxQuizTime will 'interrupt' the quiz if
// it's still going
func timeQuiz(maxQuizTime time.Duration, c chan string) {
	time.AfterFunc(maxQuizTime, func() {
		fmt.Println("\nTimer has run out!")
		c <- "timeout"
	})
}

// proctorQuiz performs the quiz, asking the user questions and gathering responses. When it finishes, it sends
// a "success" message down the channel to be returned to the calling function
func proctorQuiz(q *Quiz, reader *bufio.Reader, c chan string) {
	for index, value := range q.Questions {
		processQuestion(reader, &value, q, index)
	}
	c <- "success"
}

func proctorRandomQuiz(q *Quiz, reader *bufio.Reader, c chan string) {
	questionNumber := 1
	for len(q.Questions) > 0 {
		// gather a random index to get a random question
		randomIndex := rand.Intn(len(q.Questions))

		// store the random question
		randomQuestion := q.Questions[randomIndex]

		// remove the random question from the Questions slice
		frontSlice := q.Questions[0:randomIndex]
		backSlice := q.Questions[randomIndex+1 : len(q.Questions)]
		q.Questions = append(frontSlice, backSlice...)

		processQuestion(reader, &randomQuestion, q, questionNumber)
		questionNumber++
	}
	c <- "success"
}

func processQuestion(reader *bufio.Reader, question *Question, quiz *Quiz, questionNumber int) {
	if question.QuestionText != "" {
		fmt.Print("\n\tQuestion ", questionNumber, ":", "\t\t"+question.QuestionText+" = ")

		userAnswer, inputErr := reader.ReadString('\n')
		if inputErr == nil {
			userAnswer = strings.Replace(userAnswer, "\n", "", 1)
		} else {
			fmt.Println("An error has occurred reading user input: ", inputErr)
		}

		userAnswer = strings.TrimSpace(userAnswer)
		userAnswer = strings.ToLower(userAnswer)
		result := userAnswer == strings.ToLower(question.Answer)
		if result {
			fmt.Println("\tCorrect!")
			quiz.CorrectQuestions++
		} else {
			fmt.Println("\tIncorrect.")
		}
		question.Correct = &result
		question.UserAnswer = userAnswer
	}
}

// Results takes a Quiz and performs the processing needed to determine the % of correct answers
func (q Quiz) Results() string {
	correct, total := float64(q.CorrectQuestions), float64(q.TotalQuestions)
	ratio := 100 * (correct / total)
	percent := "%"
	return fmt.Sprintf("%2.2f%v (%v/%v) of the answers were correct.", ratio, percent, correct, total)
}
