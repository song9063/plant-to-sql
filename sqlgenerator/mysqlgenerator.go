package sqlgenerator

import (
	"fmt"
	"strings"

	"github.com/song9063/plant-to-sql/plantparser"
)

type MySQLGenerator struct {
	SchemaName     string
	OptionalString string
}

func (my *MySQLGenerator) GenerateCreateSQL(schemaName string, sqlTables []plantparser.SQLTable) (string, error) {
	if my.SchemaName == "" {
		my.SchemaName = schemaName
	}
	if my.OptionalString == "" {
		my.OptionalString = "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_unicode_ci"
	}
	arSql := make([]string, 0)

	for _, table := range sqlTables {
		sql, err := my.makeTableSQL(schemaName, table)
		if err != nil {
			return "", err
		}
		arSql = append(arSql, sql)
	}
	strSql := strings.Join(arSql, "\n\n")
	return strSql, nil
}

func (my *MySQLGenerator) makeTableSQL(schemaName string, sqlTable plantparser.SQLTable) (string, error) {
	arSql := make([]string, 0)

	// Column
	for _, col := range sqlTable.Columns {
		sql, err := my.makeColumnSQL(sqlTable.Name, col)
		if err != nil {
			return "", err
		}
		arSql = append(arSql, sql)
	}

	// Primary Key
	if len(sqlTable.PKs) > 0 {
		arSql = append(arSql, my.makePKSQL(sqlTable.Name, sqlTable.PKs))
	}

	// Unique
	for _, uq := range sqlTable.Uniques {
		sql, err := my.makeUQSQL(sqlTable.Name, uq)
		if err != nil {
			return "", err
		}
		arSql = append(arSql, sql)
	}

	// Index
	for _, index := range sqlTable.Indexes {
		sql, err := my.makeIndexSQL(sqlTable.Name, index)
		if err != nil {
			return "", err
		}
		arSql = append(arSql, sql)
	}

	// Foreign Key
	for _, fk := range sqlTable.FKs {
		sql, err := my.makeFKSql(sqlTable.Name, fk)
		if err != nil {
			return "", err
		}
		arSql = append(arSql, sql)
	}

	strOpts := my.OptionalString + " " + sqlTable.OptionalString
	strSql := strings.Join(arSql, ",\n")
	strSql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`.`%s`(\n%s", schemaName, sqlTable.Name, strSql)
	strSql = fmt.Sprintf("%s\n)%s;", strSql, strOpts)
	return strSql, nil
}

func (my *MySQLGenerator) makeFKSql(sqlTableName string, sqlFK plantparser.SQLFK) (string, error) {
	strColName := fmt.Sprintf("%s_%s", sqlFK.TargetTableName, sqlFK.TargetColumnName)
	strKeyName := fmt.Sprintf("FK_%s_%s", sqlTableName, strColName)

	strSql := fmt.Sprintf("\tCONSTRAINT `%s`", strKeyName)
	strSql = fmt.Sprintf("%s\n\t\tFOREIGN KEY (`%s`)", strSql, strColName)
	strSql = fmt.Sprintf("%s\n\t\tREFERENCES `%s`.`%s` (`%s`)", strSql, my.SchemaName, sqlFK.TargetTableName, sqlFK.TargetColumnName)

	return strSql, nil
}

func (my *MySQLGenerator) makeIndexSQL(sqlTableName string, sqlIndex plantparser.SQLIndex) (string, error) {
	arCols := make([]string, 0)
	for _, col := range sqlIndex.ColumnNames {
		strCol := strings.ReplaceAll(col, ".", "_")
		arCols = append(arCols, fmt.Sprintf("`%s`", strCol))
	}
	strSql := fmt.Sprintf("\tINDEX `%s` (%s)", sqlIndex.Name, strings.Join(arCols, ","))
	return strSql, nil
}

func (my *MySQLGenerator) makeUQSQL(sqlTableName string, sqlUQ plantparser.SQLUnique) (string, error) {
	strSql := fmt.Sprintf("\tUNIQUE INDEX `UQ_%s_%s` (`%s` ASC)", sqlTableName, sqlUQ.ColumnName, sqlUQ.ColumnName)
	return strSql, nil
}

func (my *MySQLGenerator) makePKSQL(sqlTableName string, sqlPKs []plantparser.SQLPK) string {
	arPKs := make([]string, 0)

	for _, pk := range sqlPKs {
		arPKs = append(arPKs, "`"+pk.ColumnName+"`")
	}

	return fmt.Sprintf("\tPRIMARY KEY (%s)", strings.Join(arPKs, ","))
}

func (my *MySQLGenerator) makeColumnSQL(sqlTableName string, sqlCol plantparser.SQLColumn) (string, error) {
	arSql := make([]string, 0)

	arSql = append(arSql, fmt.Sprintf("`%s` %s", sqlCol.Name, sqlCol.DataTypeString))
	if sqlCol.NotNull {
		arSql = append(arSql, "NOT NULL")
	} else {
		arSql = append(arSql, "NULL")
	}

	if sqlCol.DefaultValue != "" {
		arSql = append(arSql, fmt.Sprintf("DEFAULT %s", sqlCol.DefaultValue))
	}
	if sqlCol.AutoIncrement {
		arSql = append(arSql, "AUTO_INCREMENT")
	}

	strSql := "\t" + strings.Join(arSql, " ")
	return strSql, nil
}
