package plantparser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type TableParser struct {
	ExpStartOfEntity *regexp.Regexp // entity "bs_users" as bu {
	ExpEndOfEntity   *regexp.Regexp // }
	ExpColumn        *regexp.Regexp // *id: INT(11) UNSIGNED AI <<PK>>
	ExpFK            *regexp.Regexp // *board_settings.id <<FK>>
	ExpIndex         *regexp.Regexp // INDEX posts_for_name_reply_author_id reply_author_id

	ExpDataType *regexp.Regexp
}

func NewParser() *TableParser {
	p := TableParser{}
	p.Init()
	return &p
}

func (p *TableParser) Init() {
	p.ExpStartOfEntity, _ = regexp.Compile("(?i)^[ \\t]*entity[ \\t]+(\\\"\\w+\\\")[ \\t]*(as \\w+)?[ \\t]*{[ \\t]*$")
	p.ExpEndOfEntity, _ = regexp.Compile("^[ \\t]*}[\t]*[ \\t]*$")
	p.ExpColumn, _ = regexp.Compile("(?i)^[ \\t]*(\\*)?(\\w+)[ \\t]*:[ \\t]*(.+)[ \\t]*$")
	p.ExpFK, _ = regexp.Compile("(?i)^[ \\t]*\\*?[ \\t]*(\\w+).(\\w+)[ \\t]*<<FK>>[ \\t]*$")
	p.ExpIndex, _ = regexp.Compile("(?i)^[ \\t]*INDEX[ \\t]+([\\w., \\t]+)[ \\t]*$")

	p.ExpDataType, _ = regexp.Compile("(?i)^(INT|TINYINT|DECIMAL|VARCHAR|CHAR|TEXT|MEDIUMTEXT|DATETIME|TIMESTAMP)(\\(.+\\))?([ \\t]UNSIGNED)?([ \\t]+(DF|DEFAULT)[ \\t]+(.+))?([ \\t]+AI)?([ \\t]+<<PK>>)?([ \\t]+<<UQ>>)?")
}

// Return table names
func (p *TableParser) MakeEntityList(fileName string) ([]*PlantEntity, error) {
	entityList := make([]*PlantEntity, 0)
	entityMap := make(map[string]*PlantEntity)
	file, err := os.Open(fileName)
	if err != nil {
		return entityList, err
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fkMap := make(map[string]map[string]string) // tableName.ColumnName = typeString
	var entity *PlantEntity = nil
	for fileScanner.Scan() {
		strLine := fileScanner.Text()
		if p.ExpStartOfEntity.MatchString(strLine) {
			// Start of the entity
			ar := p.ExpStartOfEntity.FindStringSubmatch(strLine)
			strTableName := strings.ReplaceAll(ar[1], "\"", "")
			entity = NewPlantEntity(strTableName)
		} else if p.ExpEndOfEntity.MatchString(strLine) {
			// End of the entity
			if entity != nil {
				entityMap[entity.Name] = entity
				entityList = append(entityList, entity)
				entity = nil
			}
		} else if entity != nil {
			// Properties of the Entity
			if p.ExpColumn.MatchString(strLine) {
				strPlantUMLString := strings.TrimSpace(strLine)
				entity.Columns = append(entity.Columns, strPlantUMLString)
				arTokens := p.ExpColumn.FindStringSubmatch(strPlantUMLString)
				// [1]: Not Null, [2]: Name, [3]: Data Type + Constraints
				entity.ColumnMapCache[arTokens[2]] = strPlantUMLString

			} else if p.ExpFK.MatchString(strLine) {
				arTokens := p.ExpFK.FindStringSubmatch(strings.TrimSpace(strLine))
				if len(arTokens) >= 3 {
					// Make map of Foreign Key
					fkTableName := arTokens[1]
					fkTableCol := arTokens[2]
					if fkMap[fkTableName] == nil {
						fkMap[fkTableName] = make(map[string]string)
					}
					fkMap[fkTableName][fkTableCol] = ""

					entity.FKMaps = append(entity.FKMaps, map[string]string{
						fkTableName: fkTableCol,
					})

				}

			} else if p.ExpIndex.MatchString(strLine) {
				entity.Indexes = append(entity.Indexes, strings.TrimSpace(strLine))
			}

		}

	}

	if err := fileScanner.Err(); err != nil {
		return entityList, err
	}

	// Append FK to column("NotNull, Name, Type, Size, Unsigned" only)
	for _, entity := range entityMap {
		for _, fkMap := range entity.FKMaps {
			for targetTableName, targetColumnName := range fkMap {
				colCache := entityMap[targetTableName].ColumnMapCache
				plantUMLString := colCache[targetColumnName]

				arTokens := p.ExpColumn.FindStringSubmatch(plantUMLString)
				dataType := arTokens[3]
				arTypeTokens := p.ExpDataType.FindStringSubmatch(dataType)

				dataTypeString := ""
				if len(arTypeTokens) > 0 {
					dataTypeString = arTypeTokens[1]
					if len(arTypeTokens) > 1 {
						dataTypeString += arTypeTokens[2]
					}
					if len(arTypeTokens) > 2 {
						dataTypeString += arTypeTokens[3]
					}
				}
				if len(dataTypeString) < 1 {
					dataTypeString = "[[WARNING!! UNKNOWN_DATA_TYPE]]"
				}
				fkColumnString := fmt.Sprintf("*%s_%s: %s", targetTableName, targetColumnName, dataTypeString)
				entity.Columns = append(entity.Columns, fkColumnString)
			}
		}
	}

	return entityList, nil
}

func (p *TableParser) MakeSQLTable(entity *PlantEntity) (*SQLTable, error) {
	if entity == nil {
		return nil, errors.New("entity is nil.")
	}

	table := SQLTable{
		Name: entity.Name,
	}

	// Columns
	for _, col := range entity.Columns {
		sqlCol := p.makeSQLColumn(col)
		if sqlCol == nil {
			continue
		}
		table.Columns = append(table.Columns, *sqlCol)

		if sqlCol.PrimaryKey {
			table.PKs = append(table.PKs, SQLPK{ColumnName: sqlCol.Name})
		}
		if sqlCol.Unique {
			table.Uniques = append(table.Uniques, SQLUnique{ColumnName: sqlCol.Name})
		}
	}

	// Foreign Keys
	for _, fkMap := range entity.FKMaps {
		for key, val := range fkMap {
			table.FKs = append(table.FKs, SQLFK{
				TargetTableName:  key,
				TargetColumnName: val,
			})
		}
	}

	// Indexes
	for _, index := range entity.Indexes {
		arTokens := p.ExpIndex.FindStringSubmatch(index)
		if len(arTokens) != 2 {
			return nil, errors.New("Invalidate index format")
		}
		arColNames := make([]string, 0)
		for _, colName := range strings.Split(arTokens[1], ",") {
			arColNames = append(arColNames, strings.TrimSpace(colName))
		}
		sqlIndex := SQLIndex{
			ColumnNames:    arColNames,
			PlantUMLString: index,
		}
		table.Indexes = append(table.Indexes, sqlIndex)
	}

	//s, _ := json.MarshalIndent(table, "", "\t")
	//fmt.Println(string(s))

	return &table, nil
}

func (p *TableParser) makeSQLColumn(strColumn string) *SQLColumn {
	arTokens := p.ExpColumn.FindStringSubmatch(strColumn)
	//fmt.Printf("[[[%s]]]\n", strColumn)
	// [1]: Not Null, [2]: Name, [3]: Data Type + Constraints

	col := SQLColumn{
		NotNull:        len(arTokens[1]) > 0,
		Name:           arTokens[2],
		PlantUMLString: strColumn,
	}
	dataType := arTokens[3]
	//fmt.Printf("[[[---%s---]]]\n", dataType)
	arTypeTokens := p.ExpDataType.FindStringSubmatch(dataType)
	//fmt.Printf("\n\n%+v\n\n", arTypeTokens)
	// $1: Data type, $2: Size, $3: UNSIGNED
	// $6: Default value, $7: Auto increment,
	// $8: PK, $9: UQ

	if len(arTypeTokens) > 0 {
		dataTypeString := arTypeTokens[1]
		if len(arTypeTokens) > 1 {
			dataTypeString += arTypeTokens[2]
		}
		if len(arTypeTokens) > 2 {
			dataTypeString += arTypeTokens[3]
		}
		col.DataTypeString = DATA_TYPE_STRING(dataTypeString)
	}

	if len(arTypeTokens) > 6 {
		col.DefaultValue = arTypeTokens[6]
	}
	if len(arTypeTokens) > 7 {
		col.AutoIncrement = len(arTypeTokens[7]) > 0
	}
	if len(arTypeTokens) > 8 {
		col.PrimaryKey = len(arTypeTokens[8]) > 0
	}
	if len(arTypeTokens) > 9 {
		col.Unique = len(arTypeTokens[9]) > 0
	}

	return &col
}
