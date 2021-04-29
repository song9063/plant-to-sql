package plantparser

type DATA_TYPE_STRING string

const (
	DATA_TYPE_NAME_INT        DATA_TYPE_STRING = "INT"
	DATA_TYPE_NAME_TINYINT    DATA_TYPE_STRING = "TINYINT"
	DATA_TYPE_NAME_DECIMAL    DATA_TYPE_STRING = "DECIMAL"
	DATA_TYPE_NAME_VARCHAR    DATA_TYPE_STRING = "VARCHAR"
	DATA_TYPE_NAME_CHAR       DATA_TYPE_STRING = "CHAR"
	DATA_TYPE_NAME_TEXT       DATA_TYPE_STRING = "TEXT"
	DATA_TYPE_NAME_MEDIUMTEXT DATA_TYPE_STRING = "MEDIUMTEXT"
	DATA_TYPE_NAME_DATETIME   DATA_TYPE_STRING = "DATETIME"
	DATA_TYPE_NAME_TIMESTAMP  DATA_TYPE_STRING = "TIMESTAMP"
)

type SQLColumn struct {
	NotNull        bool
	AutoIncrement  bool
	PrimaryKey     bool
	Unique         bool
	Name           string
	DataTypeString DATA_TYPE_STRING
	DefaultValue   string

	PlantUMLString string
}
type SQLFK struct {
	TargetTableName  string
	TargetColumnName string
}
type SQLPK struct {
	ColumnName string
}
type SQLUnique struct {
	ColumnName string
}
type SQLIndex struct {
	ColumnNames []string

	PlantUMLString string
}
type SQLTable struct {
	Name    string
	PKs     []SQLPK
	FKs     []SQLFK
	Uniques []SQLUnique
	Columns []SQLColumn
	Indexes []SQLIndex

	OptionalString string // ex) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_unicode_ci
}

type PlantEntity struct {
	Name           string
	Columns        []string            // list of PlantUMLString
	ColumnMapCache map[string]string   // For FK. [name]=PlantUMLString
	FKMaps         []map[string]string // list of Foreign Keys(tableName=columnName)
	Indexes        []string
}

func NewPlantEntity(tableName string) *PlantEntity {
	pe := PlantEntity{
		Name: tableName,
	}
	pe.ColumnMapCache = make(map[string]string)
	return &pe
}
