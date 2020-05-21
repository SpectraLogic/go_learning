package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	//Opening and reading question file
	csvfile, err := os.Open("problems.csv")
	if err != nil {
		log.Fatalln("Couldn't open csv file ", err)
	}
	r := csv.NewReader(csvfile)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	//Asking shuffle and time limit choice
	var totalScore, totalQuestions, timeLimit, timeChoice int
	var shuffle string
	shuffleOptions := map[string]bool{"y": true, "n": false}
	fmt.Println("Do you want to shuffle the questions ? y or n")
	fmt.Scanln(&shuffle)
	if shuffleOptions[shuffle] {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(records), func(i, j int) { records[i], records[j] = records[j], records[i] })
	}
	timeOptions := map[int]int{1: 30, 2: 45, 3: 60}
	fmt.Println("Choose time limit \n 1. 30 seconds \t 2. 45 seconds \t 3. 60 seconds")
	fmt.Scanln(&timeChoice)

	//Timer begins
	timeLimit = timeOptions[timeChoice]
	timer := time.AfterFunc(time.Second*time.Duration(timeLimit), func() {
		fmt.Println("Total Score : ", totalScore)
		fmt.Println("No. of Questions : ", totalQuestions)
		os.Exit(0)
	})
	defer timer.Stop()
	for i := 0; i < len(records); i++ {
		record := records[i]
		fmt.Println("What is " + record[0] + " ?")
		userAnswer, _ := in.ReadString('\n')

		//Ignores case mismatch and additional white spaces
		if strings.Join(strings.Fields(strings.ToLower(userAnswer)), "") == strings.ToLower(record[1]) {
			totalScore++
		}
		totalQuestions++
	}

}
