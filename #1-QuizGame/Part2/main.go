package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
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
	case ans := <-ac: // get answer for current problem 'pb'
		ansi, err := strconv.Atoi(ans)
		if err != nil {
			return false
		}

		return ansi == pb.answer
	case <-time.After(timeout): // each problem get 'timeout' seconds to answer
		fmt.Println(" <-this question runs out of time")
		return false
	}
}

func parseCSV(fname string) ([][]string, error) {
	if !strings.HasSuffix(fname, "csv") {
		log.Fatalf("provided problems file '%s' is not a CSV file\n", fname)
	}

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

	c := make(chan bool)    // transport answer checking result
	go quizRound(pbs, c)

	getout := true
	correct := 0
	t := time.NewTimer(time.Duration(limit) * time.Second)

	for getout {
		select {
		case ret, ok := <-c: // collect check result
			if ok {
				if ret {
					correct++
				}
			} else { // channel c is closed
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
// check result into 'out' channel. 
func quizRound(quizs []problem, out chan<- bool) {
	var ans string
	for i, q := range quizs {
		// ask the problem ans collect answer
		fmt.Printf("Problem #%d: %s = ", i+1, q.qs)
		fmt.Scanln(&ans) // can not put this statement ahead since it will block

		// convert the answer to integer
		ansi, err := strconv.Atoi(ans)
		if err != nil {
			out <- false
			continue  // no need more comparision
		}

		// check the result
		if ansi == q.answer {
			out <- true
		} else {
			out <- false
		}
	}
	close(out)
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

	fullTime := time.Duration(limit*len(pbs)) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), fullTime)
	// after fullTime seconds, cancel() is called automatically

	answerChan := make(chan string)
	go getUserInput(ctx, answerChan)

	correct := 0
	for _, q := range pbs {
		if q.ask(time.Duration(limit)*time.Second, answerChan) {
			correct++
		}
	}

	cancel() // tell the getUserInput goroutinue to stop
	fmt.Printf("scored %d out of %d\n", correct, len(pbs))
}

func getUserInput(ctx context.Context, ac chan<- string) {
	var ans string
	for {
		select {
		case <-ctx.Done(): // cancel() is called. stop collect user input
			close(ac)
			return
		default:
			fmt.Scanln(&ans)
			ac <- ans
			ans = ""
		}
	}
}

// shuffleProblems function shuffle the problems
func shuffleProblems(pbs []problem) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < len(pbs); i++ {
		j := r.Intn(len(pbs))

		pbi, pbj := pbs[i], pbs[j]
		pbs[i], pbs[j] = pbj, pbi
	}
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

var (
	fileName  = flag.String("csv", "problems.csv", "csv file")
	timeLimit = flag.Int("limit", 30, "time limit")
	split     = flag.Bool("split", false, "split the 'limit' time in average according to the number of problems (default false)")
	shuffle   = flag.Bool("shuffle", false, "shuffle the problem or not (default false)")
)

func main() {
	flag.Parse()

	lines, err := parseCSV(*fileName)
	if err != nil {
		exit(err.Error())
	}

	allProblems, err := parseLines(lines)
	if err != nil {
		exit(err.Error())
	}

	if *shuffle {
	 	shuffleProblems(allProblems)
	}

	quiz(allProblems, *timeLimit, *split)
}
