package storage

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"errors"
	"fmt"
	"strconv"
)

const {
	ENTRY_TYPE_NIL = 0
	ENTRY_TYPE_BOOLEAN = 1
	ENTRY_TYPE_FLOAT = 2
	ENTRY_TYPE_STRING = 3

	ENTRIES_INDEX_QUERY = "CREATE INDEX path ON entries(path)"
	SELECT_ENTRIES_LIKE = "SELECT * FROM entries WHERE path LIKE ^%s/ OR path LIKE ^%s$"
}

type Entry struct {
	Id          int64  `db:"id"`
	Path        string `db:"path"`
	Value		sql.NullString `db:"value"`
	Type 		int `db:"type"`
	Owner       int64  `db:"owner"`
	Group       int64  `db:"group"`
	Permissions uint8  `db:"perm"`
}

type Database struct {
	dbMap *gorp.DbMap
}

// Db Type is either sqlite3, pg, mysql
func NewDatabase(connString string, dbName string, dbType string) (*Database, error) {
	db, connErr := sql.Open(dbType, connString)
	if connErr {
		return nil, connErr
	}

	var dialect gorp.Dialect
	switch dbType {
	case "sqlite3":
		dialect = gorp.SqliteDialect{}
	case "mysql":
		dialect = gorp.MySQLDialect{}
	case "pg":
		dialect = gorp.PostgresDialect{}
	default:
		return nil, errors.New("Unsupported db type '" + dbType + "'")
	}
	dbMap := &gorp.DbMap{
		Db:      db,
		Dialect: dialect,
	}

	dbMap.AddTableWithName(Entry{}, "entries").SetKeys(true, "Id")
	createTablesErr := dbmap.CreateTablesIfNotExists()
	if createTablesErr {
		return nil, createTablesErr
	}

	_, indexErr := dbMap.Exec(SQLITE_ENTRIES_INDEX_QUERY)
	if indexErr {
		return nil, indexErr
	}

	newDb := Database{}
	newDb.dbMap = dbMap
	return &newDb, nil
}

func (db *Database) get(path string) (map[string]interface{}, error) {
	// Get all entries "LIKE" path
	var entries []Entry
	_, selectErr := db.dbMap.Select(&entries, fmt.Sprintf(SELECT_ENTRIES_LIKE, path, path))
	if selectErr {
		return nil, selectErr
	}

	resultMap := make(map[string]interface{})
	for i, entry := range entries {
		switch entry.Type {
		case ENTRY_TYPE_NIL:
			resultMap[entry.Path] = nil
		case ENTRY_TYPE_BOOLEAN:
			if entry.Value.String == "true" {
				resultMap[entry.Path] = true
			} else {
				resultMap[entry.Path] = false
			}
		case ENTRY_TYPE_FLOAT:
			resultMap[entry.Path], _ = entry.Value.String.(float64)
		case ENTRY_TYPE_STRING:
			resultMap[entry.Path], _ = entry.Value.String
		}
	}

	return resultMap, nil
}

func (db *Database) set(values map[string]interface{}) error {
	entries := make(*Entry[], len(values), len(values))
	i := 0
	for key, value := range values {
		entry := Entry{}
		// Handle type and value first
		switch value.(type) {
		case bool:
			entry.Type = ENTRY_TYPE_BOOLEAN
			if value.(bool) {
				entry.Value = sql.NullString{
					String: "true",
					Valid: true,
				}
			} else {
				entry.Value = sql.NullString{
					String: "false",
					Valid: true,
				}
			}
		case float64:
			entry.Type = ENTRY_TYPE_FLOAT
			floatVal, err := strconv.ParseFloat(value.(float64), 64)
			if !err {
				entry.Value = sql.NullString{
					String: value,
					Valid: true,
				}
			}
		case string:
			entry.Type = ENTRY_TYPE_STRING
			entry.Value = sql.NullString{
				String: value.(string),
				Valid: true,
			}
		case nil:
			entry.Type = ENTRY_TYPE_NIL
			entry.Value = sql.NullString{
				String: "",
				Valid: false,
			}
		default:
			continue
		}
		entry.Path = key
		// TODO: owner, group, perms
		entry.Owner = 0
		entry.Group = 0
		entry.Permissions = 0

		entries[i] = &entry
		i = i + 1
	}
	// Put it in the db
	return db.dbMap.Insert(entries)
}