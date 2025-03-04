package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/alecthomas/kingpin"
)

var (
	connStr = kingpin.Arg(
		"conn", "PostgreSQL connection string in URL format").Required().String()
	schema = kingpin.Flag(
		"schema", "PostgreSQL schema name").Default("public").Short('s').String()
	title       = kingpin.Flag("title", "Diagram title").Short('T').String()
	outFile     = kingpin.Flag("output", "output file path of the uml").Short('o').String()
	moduleFile  = kingpin.Flag("module", "module file path, describing all tables for the uml module").Short('m').String()
	targetTbls  = kingpin.Flag("table", "target tables to include").Short('t').Strings()
	xTargetTbls = kingpin.Flag("exclude", "target tables to exclude").Short('x').Strings()
)

func main() {
	kingpin.Parse()

	if *moduleFile != "" {
		log.Printf("moduleFile %s", *moduleFile)

		tempTables, err := readFileAsList(*moduleFile)
		if err != nil {
			log.Fatal(err)
			return
		}
		*targetTbls = tempTables
		log.Printf("length %s", fmt.Sprint(len(*targetTbls)))
	}

	log.Println("start connecting")
	// 1. open db connection
	db, err := SqlOpenDB(*connStr)
	if err != nil {
		log.Fatal(err)
		return
	}

	ts, err := PlanterLoadTableDef(db, *schema)
	if err != nil {
		log.Fatal(err)
	}

	var tbls []*Table
	if len(*targetTbls) != 0 { // when targetTbls are specified select only these tables, otherwise select all.
		tbls = SqlFilterTables(true, ts, *targetTbls)
	} else {
		tbls = ts
	}
	if len(*xTargetTbls) != 0 { // exclude tables
		tbls = SqlFilterTables(false, tbls, *xTargetTbls)
	}
	entry, err := PlanterTableToUMLEntry(tbls)
	if err != nil {
		log.Fatal(err)
	}
	rel, err := SqlForeignKeyToUMLRelation(tbls)
	if err != nil {
		log.Fatal(err)
	}
	// ## create PlantUML-txt-file in steps:
	// 1. HEADER OF FILE
	var src []byte
	src = append([]byte("@startuml\n"))
	if len(*title) != 0 {
		src = append(src, []byte("title "+*title+"\n")...)
	}
	src = append(src, []byte("hide circle\n"+
		"skinparam linetype ortho\n")...)

	// 2. ENTRIES AND RELATIONS
	src = append(src, entry...)
	src = append(src, rel...)

	// 3. FOOTER OF FILE
	src = append(src, []byte("@enduml\n")...)

	var out io.Writer
	// when outFile specified write to file otherwise to console.
	if *outFile != "" {
		out, err = os.Create(*outFile)
		if err != nil {
			log.Fatalf("failed to create output file %s: %s", *outFile, err)
		} else {
			log.Printf("you can now execute `java -jar plantuml.jar -verbose %s`", *outFile)
		}
	} else {
		out = os.Stdout
	}
	// when writing failed, log the error.
	if _, err := out.Write(src); err != nil {
		log.Fatal(err)
	}

}

func readFileAsList(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
