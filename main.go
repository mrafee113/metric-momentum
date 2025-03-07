package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/forPelevin/gomoji"
)

type progress struct {
	id           int
	body         string
	unit         string
	count        int
	doneCount    int
	creationDate time.Time
	priority     string
	category     string
}

var (
	data         []progress
	dir          string
	taskFilepath string
)

func ensureFileExists(filepath string) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Failed to make sure that file %s exists.\n%v\n", filepath, err)
	}
	file.Close()
}

func configPath() {
	dir = os.Getenv("MEMO_DIR")
	if dir == "" {
		dir = "~/.local/share/memo"
	}
	if strings.HasPrefix(dir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalln("Failed to get user home directory path.", err)
		}
		dir = filepath.Join(homeDir, dir[2:])
	}
	taskFilepath = filepath.Join(dir, "tasks")

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalf("Failed to make sure that file %s exists.\n%v\n", dir, err)
	}
	ensureFileExists(taskFilepath)
}

func getArchiveFilepath() string {
	now := time.Now()
	return filepath.Join(dir, fmt.Sprintf(
		"archive-%d-%s", now.Year(),
		strings.ToLower(now.Format("Jan")),
	))
}

func readData(filename string) []progress {
	var data []progress
	ensureFileExists(filename)
	fileData, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file %s.\n%v\n", filename, err)
	}
	lines := strings.Split(string(fileData), "\n")
	for id, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		values := strings.Split(line, ";")
		count, err := strconv.Atoi(values[2])
		if err != nil {
			log.Fatalf("Failed to convert string '%s' to int.\n%v\n", values[2], err)
		}
		doneCount, err := strconv.Atoi(values[3])
		if err != nil {
			log.Fatalf("Failed to convert string '%s' to int.\n%v\n", values[3], err)
		}
		creationDate := time.Time{}
		err = creationDate.UnmarshalJSON([]byte(values[4]))
		if err != nil {
			log.Fatalf("Error occured trying to unmarshal creation date json.\n%s\n%v\n", values[4], err)
		}
		data = append(data,
			progress{
				id:           id,
				body:         values[0],
				unit:         values[1],
				count:        count,
				doneCount:    doneCount,
				creationDate: creationDate,
				priority:     values[5],
				category:     values[6],
			},
		)
	}
	return data
}

func writeData(filename string, data []progress) {
	var text string
	for _, prog := range data {
		jsonDate, err := prog.creationDate.MarshalJSON()
		if err != nil {
			log.Fatalf("Error occured while marshalling creation date to json.\n%v\n%v\n", prog.creationDate, err)
		}
		text += fmt.Sprintf(
			"%s;%s;%d;%d;%s;%s;%s\n",
			prog.body,
			prog.unit,
			prog.count,
			prog.doneCount,
			string(jsonDate),
			prog.priority,
			prog.category,
		)
	}
	err := os.WriteFile(filename, []byte(text), 0644)
	if err != nil {
		log.Fatalf("Failed to write to file %s.\n%v\n", filename, err)
	}
}

func barText(count, doneCount, width int) string {
	pLen := width * count / doneCount
	switch pLen {
	case 0:
		return strings.Repeat(" ", width)
	case 1:
		return ">" + strings.Repeat(" ", width-1)
	case width:
		return strings.Repeat("=", width)
	default:
		return strings.Repeat("=", pLen-1) + ">" + strings.Repeat(" ", width-pLen)
	}
}

func colorizeText(text, color, defaultColor string) string {
	if defaultColor == "" {
		defaultColor = os.Getenv("DEFAULT")
	} else {
		defaultColor = os.Getenv(defaultColor)
	}
	colorCode := os.Getenv(color)
	if colorCode == "" {
		return text
	}
	return fmt.Sprintf("%s%s%s", colorCode, text, defaultColor)
}

func formatDuration(creationDate time.Time, color bool) (result string) {
	duration := time.Since(creationDate)
	hours := duration.Hours()
	minutes := duration.Minutes()
	years := hours / 24 / 365
	months := hours / 24 / 30
	weeks := hours / 24 / 7
	days := hours / 24
	if years > 1 {
		result = fmt.Sprintf("%.1fy ", years)
	} else if months > 1 {
		result = fmt.Sprintf("%.1fmo", months)
	} else if weeks >= 2 {
		result = fmt.Sprintf("%.1fw ", weeks)
	} else if days >= 2 {
		result = fmt.Sprintf("%.0fd ", days)
	} else if hours > 1 {
		result = fmt.Sprintf("%.1fh ", hours)
	} else {
		result = fmt.Sprintf("%.0fm ", minutes)
	}
	result = fmt.Sprintf("%-5s", result)
	result = result[:5]
	if color {
		if weeks <= 1 {
			result = colorizeText(result, "WHITE", "")
		} else if weeks <= 2 {
			result = colorizeText(result, "LIGHT_ORANGE", "")
		} else if weeks < 6 {
			result = colorizeText(result, "ORANGE", "")
		} else if weeks < 12 {
			result = colorizeText(result, "RED", "")
		} else {
			result = colorizeText(result, "DARK_GREY", "")
		}
	}
	return result
}

func usage() {
	version, err := os.ReadFile("version")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, `
Metric Momentum %s
Usage: %s <command> [options]

	Commands:
		print: Display all progress entries
		create: Creates a new progress entry
		delete: Delete and archive a progress entry
		modify: Modify an existing progress entry
		inc: increment selected entry count
		dec: decrement selected entry count
		echo: print an entry attribute
		
		- use '<command> -h' for more info

	Environment Variables:
		MEMO_DIR: designates the directory within which the files are stored. Defaults to '~/.local/share/memo'
		MEMO_WIDTH: sets the width of the output. Defaults to 20.`,
		version, os.Args[0],
	)
}

func cmdPrint() {
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	colorize := fs.Bool("color", false, "Apply colors tailored for Conky")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s print [options]\n\n", os.Args[0])
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	data = readData(taskFilepath)
	widthStr := os.Getenv("MEMO_WIDTH")
	if widthStr == "" {
		widthStr = "10"
	}
	width, err := strconv.Atoi(widthStr)
	if err != nil {
		panic(err)
	}

	slices.SortFunc(data, func(a, b progress) int {
		if a.category == "" && b.category != "" {
			return 1
		} else if a.category != "" && b.category == "" {
			return -1
		} else if a.category < b.category {
			return -1
		} else if a.category > b.category {
			return 1
		}

		if a.priority != "" && b.priority == "" {
			return -1
		} else if a.priority == "" && b.priority != "" {
			return 1
		} else if a.priority < b.priority {
			return -1
		} else if a.priority > b.priority {
			return 1
		}
		aBody, bBody := gomoji.RemoveEmojis(a.body), gomoji.RemoveEmojis(b.body)
		aBody, bBody = strings.ToLower(aBody), strings.ToLower(bBody)
		aBody, bBody = strings.TrimSpace(aBody), strings.TrimSpace(bBody)

		if aBody < bBody {
			return -1
		} else if aBody > bBody {
			return 1
		} else {
			return 0
		}
	})
	lineMax := slices.MaxFunc(data, func(a, b progress) int {
		if len(a.unit)+len(a.priority)+len(a.body) >
			len(b.unit)+len(b.priority)+len(b.body) {
			return 1
		} else {
			return -1
		}
	})
	lineMaxSize := 34 + 4 + len(lineMax.unit) + len(lineMax.priority) + len(lineMax.body)
	curCategory := "noway"
	getCategoryLine := func(category string, size int) string {
		var prefix string
		if len(category) > 0 {
			prefix = fmt.Sprintf("> %s ", category)
		} else {
			prefix = ""
		}
		return prefix + strings.Repeat("—", size-len(prefix))
	}
	var lines []string
	for _, prog := range data {
		id := prog.id
		body, unit, count, doneCount := prog.body, prog.unit, prog.count, prog.doneCount
		bar_text := barText(prog.count, prog.doneCount, width)
		percentage := 100 * prog.count / prog.doneCount
		durationText := formatDuration(prog.creationDate, false)
		priority := prog.priority
		if priority != "" {
			priority = fmt.Sprintf("(%s)", priority)
		}
		var (
			percentageText = fmt.Sprintf("%3d%%", percentage)
			countText      = fmt.Sprintf("%3d", count)
			doneCountText  = fmt.Sprintf("%3d", doneCount)
			idText         = fmt.Sprintf("%-2d", id)
		)

		var categoryLine string
		if prog.category != curCategory || curCategory == "noway" {
			curCategory = prog.category
			categoryLine = "-"
		}
		if categoryLine != "" {
			categoryLine = getCategoryLine(curCategory, lineMaxSize)
		}
		if *colorize {
			if categoryLine != "" {
				categoryLine = colorizeText(categoryLine, "BRIGHT_RED", "WHITE")
			}
			var defaultColor string
			if priority != "" {
				switch priority[1] {
				case 'A':
					priority = colorizeText(priority, "PRI_A", "")
					body = colorizeText(body, "PRI_A", "")
					defaultColor = "PRI_A"
				case 'B':
					priority = colorizeText(priority, "PRI_B", "")
					body = colorizeText(body, "PRI_B", "")
					defaultColor = "PRI_B"
				case 'C':
					priority = colorizeText(priority, "PRI_C", "")
					body = colorizeText(body, "PRI_C", "")
					defaultColor = "PRI_C"
				case 'D':
					priority = colorizeText(priority, "PRI_D", "")
					body = colorizeText(body, "PRI_D", "")
					defaultColor = "PRI_D"
				}
			}
			bodySlice := strings.Split(body, " ")
			for ndx := 0; ndx < len(bodySlice); ndx++ {
				if len(bodySlice[ndx]) < 2 {
					continue
				}
				switch bodySlice[ndx][0] {
				case '+':
					bodySlice[ndx] = colorizeText(bodySlice[ndx], "COLOR_PLUS", defaultColor)
				case '-':
					bodySlice[ndx] = colorizeText(bodySlice[ndx], "COLOR_DASH", defaultColor)
				case '@':
					bodySlice[ndx] = colorizeText(bodySlice[ndx], "COLOR_ATSIGN", defaultColor)
				}
			}
			body = strings.Join(bodySlice, " ")
			body = gomoji.ReplaceEmojisWithFunc(body, func(em gomoji.Emoji) string {
				return fmt.Sprintf("${font2}%s${font}", em.Character)
			})

			idText = colorizeText(idText, "COLOR_NUMBER", "")
			if 0 <= percentage && percentage < 20 {
				bar_text = colorizeText(bar_text, "PRI_5", "")
				percentageText = colorizeText(percentageText, "PRI_5", "")
			} else if 20 <= percentage && percentage < 40 {
				bar_text = colorizeText(bar_text, "PRI_4", "")
				percentageText = colorizeText(percentageText, "PRI_4", "")
			} else if 40 <= percentage && percentage < 60 {
				bar_text = colorizeText(bar_text, "PRI_3", "")
				percentageText = colorizeText(percentageText, "PRI_3", "")
			} else if 60 <= percentage && percentage < 80 {
				bar_text = colorizeText(bar_text, "PRI_2", "")
				percentageText = colorizeText(percentageText, "PRI_2", "")
			} else if 80 <= percentage && percentage <= 100 {
				bar_text = colorizeText(bar_text, "PRI_1", "")
				percentageText = colorizeText(percentageText, "PRI_1", "")
			}
			doneCountText = colorizeText(doneCountText, "COLOR_DONE", "")

			durationText = formatDuration(prog.creationDate, true)
		}

		if categoryLine != "" {
			lines = append(lines, categoryLine)
		}
		if priority != "" {
			priority += " "
		}
		line := fmt.Sprintf(
			"%s %s/%s(%s) %s %s (%s) %s%s",
			idText, countText, doneCountText,
			percentageText, bar_text,
			durationText, unit, priority, body,
		)
		// line = fmt.Sprintf("%d+%d=%d -- %s", prefixCount, postfixCount, totalCount, line)
		lines = append(lines, line)
	}
	fmt.Println(strings.Join(lines, "\n"))
}

func cmdCreate() {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	body := fs.String("body", "", "Text describing the progress (required)")
	unit := fs.String("unit", "", "Unit of progress measurement (required)")
	count := fs.Int("count", 0, "Initial progress value (defaults to 0)")
	doneCount := fs.Int("doneCount", 0, "Target completion value (required)")
	priority := fs.String("priority", "", "Priority of the progress (defaults to empty string)")
	category := fs.String("category", "", "Category in which the progress belongs to (defaults to empty string)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s create -body <body> -unit <unit>"+
			"-doneCount <doneCount> [-count <count>] [-priority <priority>]"+
			"[-category <category>]\n\n", os.Args[0])
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	if *body == "" || *unit == "" || *doneCount == 0 {
		fs.Usage()
		log.Fatalln("One of the required fields were left out.")
	}
	data = readData(taskFilepath)
	data = append(data, progress{
		body:         *body,
		unit:         *unit,
		count:        *count,
		doneCount:    *doneCount,
		creationDate: time.Now(),
		priority:     *priority,
		category:     *category,
	})
	writeData(taskFilepath, data)
}

func archive(id int) {
	data = readData(taskFilepath)
	if id == -1 || id >= len(data) {
		log.Fatalln("Either the given id entry does not exist or it was not given.")
	}
	entry := data[id]
	data = append(data[:id], data[id+1:]...)
	writeData(taskFilepath, data)

	archiveFilepath := getArchiveFilepath()
	archiveData := readData(archiveFilepath)
	archiveData = append(archiveData, entry)
	sort.Slice(archiveData, func(i, j int) bool {
		return archiveData[i].creationDate.Before(data[j].creationDate)
	})
	writeData(archiveFilepath, archiveData)
}

func cmdDelete() {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	id := fs.Int("id", -1, "Id of the progress entry to delete (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s delete -id <id>\n\n", os.Args[0])
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])
	archive(*id)
}

func cmdModify() {
	fs := flag.NewFlagSet("modify", flag.ExitOnError)
	id := fs.Int("id", -1, "Id of the progress entry to modify (required)")
	body := fs.String("body", "", "Text describing the progress")
	unit := fs.String("unit", "", "Unit of progress measurement")
	count := fs.Int("count", 0, "Initial progress value (defaults to 0)")
	doneCount := fs.Int("doneCount", 0, "Target completion value")
	priority := fs.String("priority", "noway", "Priority of the progress")
	category := fs.String("category", "noway", "Category of the progress")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s modify -id <id> [-name <name>] [-unit <unit>]"+
			"[-count <count>] [-doneCount <doneCount>] [-priority <priority>] [-category <category>]\n "+
			"At least one optional arg is required.\n\n", os.Args[0])
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	data = readData(taskFilepath)
	if *id == -1 || *id >= len(data) {
		fs.Usage()
		log.Fatalln("Required id was left out.")
	}
	if *body == "" && *unit == "" && *count == 0 && *doneCount == 0 && *priority == "noway" && *category == "noway" {
		fs.Usage()
		log.Fatalln("None of the flags were selected.")
	}
	prog := &data[*id]
	if *body != "" {
		prog.body = *body
	}
	if *unit != "" {
		prog.unit = *unit
	}
	if *count != 0 {
		prog.count = *count
	}
	if *doneCount != 0 {
		prog.doneCount = *doneCount
	}
	if *priority != "noway" {
		prog.priority = *priority
	}
	if *category != "noway" {
		prog.category = *category
	}
	writeData(taskFilepath, data)
}

func increment(id, count int) {
	data = readData(taskFilepath)
	if id < 0 || id >= len(data) {
		log.Fatalln("The given id does not exist.")
	}
	data[id].count += count
	if data[id].count >= data[id].doneCount {
		archive(id)
	}
	writeData(taskFilepath, data)
}

func cmdInc() {
	fs := flag.NewFlagSet("inc", flag.ExitOnError)
	id := fs.Int("id", -1, "Id of the progress entry to increment (required)")
	count := fs.Int("count", 1, "Value of the addition (defaults to 1)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s inc -id <id> [-count <count>]\n\n", os.Args[0])
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])
	increment(*id, *count)
}

func decrement(id, count int) {
	data = readData(taskFilepath)
	if id < 0 || id >= len(data) {
		log.Fatalln("The given id does not exist.")
	}
	data[id].count = max(0, data[id].count-count)
	writeData(taskFilepath, data)
}

func cmdDec() {
	fs := flag.NewFlagSet("inc", flag.ExitOnError)
	id := fs.Int("id", -1, "Id of the progress entry to decrement (required)")
	count := fs.Int("count", 1, "Value of the subtraction (defaults to 1)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s inc -id <id> [-count <count>]\n\n", os.Args[0])
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])
	decrement(*id, *count)
}

func cmdEcho() {
	fs := flag.NewFlagSet("echo", flag.ExitOnError)
	id := fs.Int("id", -1, "Id of the progress entry to echo (required)")
	body := fs.Bool("body", false, "Text describing the progress")
	unit := fs.Bool("unit", false, "Unit of progress measurement")
	count := fs.Bool("count", false, "Current progress value")
	doneCount := fs.Bool("doneCount", false, "Target completion value")
	priority := fs.Bool("priority", false, "Priority of the progress")
	category := fs.Bool("category", false, "Category of the progress")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s echo -id <id> [-body] [-unit] "+
			"[-doneCount] [-count] [-priority] [-category]\n "+
			"One and only one argument besides id should be selected.\n\n", os.Args[0])
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	data = readData(taskFilepath)
	if *id == -1 || *id >= len(data) {
		fs.Usage()
		log.Fatalln("Required id was left out.")
	}
	trueCount := 0
	for _, val := range []bool{*body, *unit, *count, *doneCount, *priority, *category} {
		if val {
			trueCount++
		}
	}
	if trueCount != 1 {
		log.Fatalln("One, and only one flag besides id should've been set.")
	}

	if *body {
		fmt.Println(data[*id].body)
	} else if *unit {
		fmt.Println(data[*id].unit)
	} else if *count {
		fmt.Println(data[*id].count)
	} else if *doneCount {
		fmt.Println(data[*id].doneCount)
	} else if *priority {
		fmt.Println(data[*id].priority)
	} else if *category {
		fmt.Println(data[*id].category)
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}
	operation := os.Args[1]
	configPath()

	switch operation {
	case "print":
		cmdPrint()
	case "create":
		cmdCreate()
	case "delete":
		cmdDelete()
	case "modify":
		cmdModify()
	case "inc":
		cmdInc()
	case "dec":
		cmdDec()
	case "echo":
		cmdEcho()
	default:
		usage()
		os.Exit(1)
	}
}
