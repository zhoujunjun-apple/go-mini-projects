package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type quiz struct {
	qs     string
	answer int
}

// parseProblemCsv function read all the problems from file 'fp',
// which should located at current working directory.
func parseProblemCsv(fp string) ([]quiz, error) {
	f, err := os.Open(filepath.Join(".", fp))
	defer f.Close()
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)

	ret := make([]quiz, 0)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		ans, err := strconv.Atoi(strings.TrimSpace(record[1]))
		if err != nil {
			return nil, err
		}

		q := quiz{
			qs:     record[0],
			answer: ans,
		}

		ret = append(ret, q)
	}

	return ret, nil
}

func quizMain(qs []quiz) error {
	n := len(qs)
	if n <= 0 {
		return fmt.Errorf("no valid problems found")
	}

	correct := 0
	for i := 1; i <= n; i++ {
		q := qs[i-1]

		fmt.Printf("Problem #%d: %s = ", i, q.qs)

		// get console input with fmt.Scanln() function
		var ans string
		fmt.Scanln(&ans)

		ansi, err := strconv.Atoi(ans)
		if err != nil {
			return err
		}

		if ansi == q.answer {
			correct++
		}
	}

	fmt.Printf("You scored %d out of %d.\n", correct, n)
	return nil
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func main() {
	csvFile := flag.String("csv", "problems.csv", "a csv file in the format of 'question,answer'")
	flag.Parse()

	quizs, err := parseProblemCsv(*csvFile)
	if err != nil {
		exit(err.Error())
	}

	err = quizMain(quizs)
	if err != nil {
		exit(err.Error())
	}
}
