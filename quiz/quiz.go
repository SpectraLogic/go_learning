/*
	quiz defines a quiz and its questions and provides methods for starting and processing quizzes as well as reading in
    a Quiz from file.
*/
package quiz

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
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

// New creates a new Quiz struct based on the information in the file at the given file name
func New(filename string) (*Quiz, error) {
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

	var questions []Question
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

	return &Quiz{
		Questions:        questions,
		Filename:         filename,
		TotalQuestions:   len(questions),
		CorrectQuestions: 0,
	}, err
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

	completed := make(chan string)

	if random {
		go proctorRandomQuiz(q, reader, completed)
	} else {
		go proctorQuiz(q, reader, completed)
	}

	timeout := time.NewTimer(maxQuizTime)
	defer timeout.Stop()

	select {
	case <-timeout.C:
		return fmt.Errorf("the quiz timer (%v seconds) ran out before the user completed the quiz",
			maxQuizTime)
	case <-completed:
		return nil
	}
}

// proctorQuiz performs the quiz, asking the user questions and gathering responses. When it finishes, it sends
// a "success" message down the channel to be returned to the calling function
func proctorQuiz(q *Quiz, reader *bufio.Reader, c chan string) {
	for index, value := range q.Questions {
		processQuestion(reader, &value, q, index)
	}
	c <- "done"
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
	c <- "done"
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
