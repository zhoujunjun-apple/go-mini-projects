package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type problem struct {
	qs     string
	answer int
}

func parseCSV(fname string) ([][]string, error) {
	f, err := os.Open(filepath.Join(".", fname))
	if err != nil {
		return nil, fmt.Errorf("Error occured when opening file %s : %s", fname, err.Error())
	}
	defer f.Close()

	r := csv.NewReader(f)
	return r.ReadAll()
}

// parseLines function convert general [][]string to []problem 
func parseLines(lines [][]string) ([]problem, error) {
	ret := make([]problem, len(lines)) // we already know the length of quiz slice

	for i, line := range lines {
		ans, err := strconv.Atoi(strings.TrimSpace(line[1]))
		if err != nil {
			return nil, err
		}

		ret[i] = problem{
			qs:     line[0],
			answer: ans,
		}
	}
	return ret, nil
}

// hint function give hint to start this quiz
func hint(limit int) {
	fmt.Printf("You have %d seconds to finish this quiz, press [Y] to start: ", limit)

	var s string
	for {
		_, err := fmt.Scanln(&s)
		if err == nil && s == "Y" {
			break
		}
		fmt.Printf("press [Y] to start: ")
	}
}

// startQuiz function is the main function to start quizing
func startQuiz(pbs []problem, limit int) {
	hint(limit)
	fmt.Println("quiz is started ...")

	c := make(chan int8)
	go quizRound(pbs, c)

	getout := true
	correct := 0
	t := time.NewTimer(time.Duration(limit) * time.Second)
	for getout {
		select {
		case ret := <-c: // collect check result
			if ret >= 0 {
				correct += int(ret)
			} else {
				// quiz is finished
				fmt.Println("\nYou've finished all the problems!")
				getout = false  
				// can not use break, since its only work for select, not for 'for' loop
			}
		case <-t.C: // the global quiz timer
			// quiz time is expired
			fmt.Println("\ntime is up!")
			getout = false
		}
	}

	fmt.Printf("scored %d out of %d\n", correct, len(pbs))
}

// quizRound function ask each question in quizs, and put the 
// check result into 'out' channel
func quizRound(quizs []problem, out chan<- int8) {
	var ans string
	for i, q := range quizs {
		fmt.Printf("Problem #%d: %s = ", i+1, q.qs)
		fmt.Scanln(&ans) // can not put this statement ahead since it will block
		ansi, _ := strconv.Atoi(ans)

		if ansi == q.answer {
			out <- 1 // 1 is right
		} else {
			out <- 0 // 0 is wrong or empty answer
		}
	}
	out <- -1 // -1 is the end of quiz
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func main() {
	fileName := flag.String("csv", "problems.csv", "csv file")
	timeLimit := flag.Int("limit", 30, "time limit")
	flag.Parse()

	lines, err := parseCSV(*fileName)
	if err != nil {
		exit(err.Error())
	}

	allProblems, err := parseLines(lines)
	if err != nil {
		exit(err.Error())
	}

	startQuiz(allProblems, *timeLimit)
}
