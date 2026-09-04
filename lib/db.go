package lib

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

type RunnableQuery func() (sql.Result, error)
type RunnableSelectQuery func() (*sqlx.Rows, error)

func BuildInsertQuery(
	connection *sqlx.DB,
	table string,
	columns map[string]interface{},
) RunnableQuery {
	args := []interface{}{}
	namePart := []string{}
	valuePart := []string{}
	for key, value := range columns {
		namePart = append(namePart, fmt.Sprintf("`%s`", key))
		valuePart = append(valuePart, fmt.Sprintf("$%d", len(args)+1))
		args = append(args, value)
	}

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, strings.Join(namePart, ","), strings.Join(valuePart, ","))
	log.Printf("%s %v", query, args)

	return func() (sql.Result, error) {
		return connection.Exec(query, args...)
	}
}

func BuildSelectStatement(
	connection *sqlx.DB,
	table string,
	columns []string,
	where map[string]interface{},
) RunnableSelectQuery {
	columnPart := []string{}
	for _, column := range columns {
		columnPart = append(columnPart, fmt.Sprintf("`%s`", column))
	}

	wherePart := []string{}
	args := []interface{}{}
	for key, value := range where {
		wherePart = append(wherePart, fmt.Sprintf("`%s`=$%d", key, len(args)+1))
		args = append(args, value)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.Join(columnPart, ","), table, strings.Join(wherePart, " AND "))
	log.Printf("%s %v", query, args)

	return func() (*sqlx.Rows, error) {
		return connection.Queryx(query, args...)
	}
}

func GetMaxId(
	connection *sqlx.DB,
	table string,
	column string,
) (int, error) {
	result, err := connection.Queryx(fmt.Sprintf("SELECT MAX(`%s`) AS max_id FROM `%s`", column, table))
	if err != nil {
		return 0, fmt.Errorf("Cannot query max_id: %w", err)
	}
	defer result.Close()

	if !result.Next() {
		return 0, nil
	}

	row := struct {
		MaxId int `db:"max_id"`
	}{}
	err = result.StructScan(&row)
	if err != nil {
		return 0, fmt.Errorf("Cannot map max_id: %w", err)
	}

	return row.MaxId, nil
}
