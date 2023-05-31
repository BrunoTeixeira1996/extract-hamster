package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Data struct {
	Activity string
	Category string
	Range    map[string]string
	Duration float64
}

type Date struct {
	Year   int
	Month  time.Month
	Day    int
	Hour   int
	Minute int
	Sec    int
}

type Output struct {
	Data      Data
	TotalTime float64
}

// Split and validate dates
func checkDates(dates string) error {
	splittedDates := strings.Split(dates, " ")
	isDate := func(date string) error {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return fmt.Errorf("Error parsing the dates: %w", err)
		}
		return nil
	}

	err1 := isDate(splittedDates[0])
	err2 := isDate(splittedDates[1])
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

// Correctly assigns the date
func (d *Date) assignDate(c string) error {
	var (
		res int
		err error
		aux []int
	)

	// [2023 03 01 09:02]
	temp := strings.Split(strings.Join(strings.Split(c, "-"), " "), " ")
	count := 0

	// Converts everything to ints
	convert := func(c string) (int, error) {
		if res, err = strconv.Atoi(c); err != nil {
			log.Printf("Error while converting %s to int:%s\n", c, err)
			return 0, err
		}
		return res, nil
	}

	for _, i := range temp {
		if count == 3 {
			//split the hour and minutes
			hourAndMinutes := strings.Split(i, ":")
			if d.Hour, err = convert(hourAndMinutes[0]); err != nil {
				return err
			}

			if d.Minute, err = convert(hourAndMinutes[1]); err != nil {
				return err
			}

		} else {
			if res, err = convert(i); err != nil {
				return err
			}
			aux = append(aux, res)
		}

		count++
	}

	d.Year, d.Month, d.Day = aux[0], time.Month(aux[1]), aux[2]

	return nil
}

// Calcs durations and returns to minutes
func (d *Data) calcDurationInMinutes() error {
	var (
		err    error
		secs   time.Duration
		startT Date
		endT   Date
	)

	if err = startT.assignDate(d.Range["start"]); err != nil {
		return err
	}

	if err = endT.assignDate(d.Range["end"]); err != nil {
		return err
	}

	firstDate := time.Date(startT.Year, startT.Month, startT.Day, startT.Hour, startT.Minute, startT.Sec, 0, time.UTC)
	secondDate := time.Date(endT.Year, endT.Month, endT.Day, endT.Hour, endT.Minute, endT.Sec, 0, time.UTC)
	difference := secondDate.Sub(firstDate)

	if secs, err = time.ParseDuration(difference.String()); err != nil {
		return fmt.Errorf("Error while parsing duration:%w", err)
	}

	d.Duration = secs.Minutes()

	return nil
}

// Function that handles the errors
func run() error {
	outFlag := flag.Bool("out", false, "-out gives correct formating to copy and paste into libre cal or excel")
	rangeFlag := flag.String("range", "", "'2023-03-01  2023-03-31'")
	calcMinutesFlag := flag.Bool("calc-minutes", false, "-calc-minutes to show duration in minutes of outputed range")
	projectFlag := flag.String("project", "", "-project 'PROJECT_CODE'")
	flag.Parse()

	// 2023-03-01  2023-03-31 has 21 chars
	if len(*rangeFlag) < 21 {
		return fmt.Errorf("Please provide the timestamp range like -range '2023-03-01  2023-03-31'")
	}

	if err := checkDates(*rangeFlag); err != nil {
		return err
	}

	params := []string{"call", "--session", "--dest",
		"org.gnome.Hamster", "--object-path",
		"/org/gnome/Hamster", "--method",
		"org.gnome.Hamster.GetFactsJSON", *rangeFlag, "",
	}

	// Calls gdbus
	out, err := exec.Command("gdbus", params...).Output()
	if err != nil {
		return err
	}

	cleanString := func(r rune) bool {
		return r == '(' || r == ')' || r == '\''
	}

	// Clean output
	cleanedOutput := strings.FieldsFunc(string(out), cleanString)

	// Create valid json
	jsonObject := strings.TrimSuffix(strings.Replace(strings.Join(cleanedOutput, ""), "\n", "", -1), ",")

	var data []Data

	err = json.Unmarshal([]byte(jsonObject), &data)
	if err != nil {
		return err
	}

	var t float64
	var output []Output

	for _, d := range data {
		var temp Output

		d.calcDurationInMinutes()

		// If the project flag is set and is equal do the category or if the project flag is not set build the Output struct
		if len(*projectFlag) > 0 && *projectFlag == d.Category || len(*projectFlag) == 0 {
			temp = Output{
				Data: Data{
					Activity: d.Activity,
					Category: d.Category,
					Range: map[string]string{
						"start": d.Range["start"],
						"end":   d.Range["end"],
					},
					Duration: d.Duration,
				},
			}

			if *calcMinutesFlag {
				t += d.Duration
			}

			// This means that we didn't set the projectFlag or the projectFlag != d.Category, so we continue to the next iteration
		} else {
			continue
		}

		if *calcMinutesFlag {
			temp.TotalTime = t
		}

		output = append(output, temp)
	}

	if len(output) == 0 {
		return fmt.Errorf("Looks like that project does not exist")
	}

	// Output according to the flags passed
	if *outFlag {
		for _, d := range output {
			fmt.Printf("%s,%s,%s,%s\n", d.Data.Activity, d.Data.Range["start"], d.Data.Range["end"], d.Data.Category)
		}
	}

	// Extract minutes from current search
	if *calcMinutesFlag {
		fmt.Printf("Total time: %.f (minutes)\n", output[len(output)-1].TotalTime)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
