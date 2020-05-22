package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Quiz struct {
	filename       string
	timeLimit      int
	shuffle        bool
	problems       [][]string
	totalScore     int
	totalQuestions int
}

func (q *Quiz) readProblems() {
	csvfile, err := os.Open(q.filename)
	if err != nil {
		log.Fatalln("Couldn't open csv file ", err)
	}
	r := csv.NewReader(csvfile)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}
	q.problems = records
}
func (q *Quiz) shuffleProblems() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(q.problems), func(i, j int) { q.problems[i], q.problems[j] = q.problems[j], q.problems[i] })
}

func (q *Quiz) printProblems() {
	in := bufio.NewReader(os.Stdin)
	for i := 0; i < len(q.problems); i++ {
		question, answer := q.problems[i][0], q.problems[i][1]
		fmt.Println("What is " + question + " ?")
		userAnswer, _ := in.ReadString('\n')

		//Ignores case mismatch and additional white spaces
		if strings.Join(strings.Fields(strings.ToLower(userAnswer)), "") == strings.ToLower(answer) {
			q.totalScore++
		}
		q.totalQuestions++
	}
}

func main() {

	//Get flags values
	filename := flag.String("filename", "problems.csv", "File containing the questions")
	timeLimit := flag.Int("timeLimit", 30, "Time limit in seconds")
	shuffle := flag.Bool("shuffle", false, "Shuffle questions")
	flag.Parse()

	q := Quiz{filename: *filename, timeLimit: *timeLimit, shuffle: *shuffle}

	q.readProblems()
	if q.shuffle {
		q.shuffleProblems()
	}

	//Setting timer
	timer := time.AfterFunc(time.Second*time.Duration(q.timeLimit), func() {
		fmt.Println("Total Score : ", q.totalScore)
		fmt.Println("No. of Questions : ", q.totalQuestions)
		os.Exit(0)
	})
	defer timer.Stop()

	q.printProblems()

}
