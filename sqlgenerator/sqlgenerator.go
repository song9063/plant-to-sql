package sqlgenerator

import "github.com/song9063/plant-to-sql/plantparser"

type SQLGeneratorInterface interface {
	GenerateCreateSQL(sqlTable []plantparser.SQLTable) (string, error)
}
