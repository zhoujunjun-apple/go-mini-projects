package main

import (
	"context"
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

// ask function show the current problem and check the result
func (pb *problem) ask(timeout time.Duration, ac <-chan string) bool {
	fmt.Printf("Problem: %s = ", pb.qs)

	select {
	case ans := <-ac:
		ansi, _ := strconv.Atoi(ans)
		return ansi == pb.answer
	case <-time.After(timeout):
		fmt.Println("this question runs out of time")
		return false
	}
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
func hint(msg string) {
	fmt.Printf(msg)

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
	hint(fmt.Sprintf("You have %d seconds to finish the full quiz, press [Y] to start: ", limit))
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

// splitTime function split 'limit' time into 'n' part in averate
func splitTime(limit, n int) int {
	var ret int
	ret = limit / n
	if ret < 2 { // should not less than 2
		ret = 2
	}
	return ret
}

// quiz function is the main entry
func quiz(bps []problem, limit int, split bool) {
	if split {
		timeout := splitTime(limit, len(bps))
		averageQuiz(bps, timeout)
	} else {
		startQuiz(bps, limit)
	}
}

// averageQuiz function give each 'problem' 'limit' seconds
func averageQuiz(pbs []problem, limit int) {
	hint(fmt.Sprintf("You have %d seconds to solve each problem, press [Y] to start: ", limit))

	answerChan := make(chan string)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(limit*len(pbs))*time.Second)

	go getUserInput(ctx, answerChan)

	correct := 0
	for _, q := range pbs {
		if q.ask(time.Duration(limit)*time.Second, answerChan) {
			correct++
		}
	}

	cancel()
	fmt.Printf("scored %d out of %d\n", correct, len(pbs))
}

func getUserInput(ctx context.Context, ac chan<- string) {
	var ans string
	for {
		select {
		case <-ctx.Done():
			close(ac)
			return
		default:
			fmt.Scanln(&ans)
			ac <- ans
			ans = ""
		}
	}
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func main() {
	fileName := flag.String("csv", "problems.csv", "csv file")
	timeLimit := flag.Int("limit", 30, "time limit")
	split := flag.Bool("split", false, "split the 'limit' time in average according to the number of problems (default false)")
	flag.Parse()

	lines, err := parseCSV(*fileName)
	if err != nil {
		exit(err.Error())
	}

	allProblems, err := parseLines(lines)
	if err != nil {
		exit(err.Error())
	}

	quiz(allProblems, *timeLimit, *split)
}
