package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/song9063/plant-to-sql/plantparser"
	"github.com/song9063/plant-to-sql/sqlgenerator"
)

func main() {

	inputFileName := flag.String("in", "", "Input filename (ex: -in mytables.plantuml)")
	schemaName := flag.String("s", "defaultschema", "Schema name (ex: -s mydb)")
	flag.Parse()

	if *inputFileName == "" {
		flag.Usage()
		return
	}

	parser := plantparser.NewParser()
	entityMap, err := parser.MakeEntityList(*inputFileName)
	if err != nil {
		log.Fatal(err)
		return
	}

	arTables := make([]plantparser.SQLTable, 0)
	for _, entity := range entityMap {
		table, err := parser.MakeSQLTable(entity)
		if err != nil {
			log.Fatal(err)
		}
		arTables = append(arTables, *table)
	}

	sqlGenerator := &sqlgenerator.MySQLGenerator{}
	strSql, err := sqlGenerator.GenerateCreateSQL(*schemaName, arTables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strSql)
}
