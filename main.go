package main

import (
	"flag"
	"log"
	"os"
)

type progress struct {
	name       string
	unit       string
	startCount int
	doneCount  int
}

func cmd() map[string]interface{} {
	flags := make(map[string]interface{})
	cwd, _ := os.Getwd()

	flags["filename"] = flag.String("filename", cwd, "The file in which data is stored.")
	flags["operation"] = flag.String("operation", "print", "What the app should do.")

	createSet := flag.NewFlagSet("create", flag.ExitOnError)
	flags["createName"] = createSet.String("name", "", "The name of the progression.")
	flags["createUnit"] = createSet.String("unit", "", "The unit of the progression.")
	flags["createStartCount"] = createSet.Int("start-count", 0, "The starting value for counting.")
	flags["createdoneCount"] = createSet.Int("done-count", 0, "The completion value for counting.")

	// printSet := flag.NewFlagSet("print", flag.ExitOnError)
	// just one op flag. don't do named shit...

	deleteSet := flag.NewFlagSet("delete", flag.ExitOnError)
	flags["deleteName"] = deleteSet.String("name", "", "The name of the progression.")

	if len(os.Args) < 2 {
		log.Println("Expected 'create', 'print', or 'delete' subcommands.")
		os.Exit(1)
	}
}

func main() {
}
