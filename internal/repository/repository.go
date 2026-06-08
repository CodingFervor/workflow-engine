package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/CodingFervor/workflow-engine/internal/database"
	"github.com/CodingFervor/workflow-engine/pkg/logger"
	"github.com/CodingFervor/workflow-engine/pkg/pagination"
)

// Base provides generic CRUD helpers.
type Base struct{}

func (b *Base) DB() *sql.DB { return database.DB }

func (b *Base) List(query string, args []interface{}, scan func(*sql.Rows) (interface{}, error)) ([]interface{}, error) {
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []interface{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

func (b *Base) GetByID(table string, id int64, scan func(*sql.Row) (interface{}, error)) (interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", table)
	row := database.DB.QueryRow(query, id)
	return scan(row)
}

func (b *Base) Count(table string, where string, args []interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if where != "" {
		query += " WHERE " + where
	}
	var count int64
	if err := database.DB.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (b *Base) Paginate(table string, p pagination.Params, where string, args []interface{}, scan func(*sql.Rows) (interface{}, error)) (*pagination.Result, error) {
	var countArgs []interface{}
	count, err := b.Count(table, where, countArgs)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT * FROM %s", table)
	if where != "" {
		query += " WHERE " + where
	}
	query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, p.Limit(), p.Offset())
	items, err := b.List(query, args, scan)
	if err != nil {
		return nil, err
	}
	result := p.Result(items, count)
	return &result, nil
}

func (b *Base) SoftDelete(table string, id int64) error {
	query := fmt.Sprintf("UPDATE %s SET status = 'deleted', updated_at = NOW() WHERE id = $1", table)
	_, err := database.DB.Exec(query, id)
	if err != nil {
		logger.Error("soft delete failed", "table", table, "id", id, "error", err)
	}
	return err
}

func (b *Base) HardDelete(table string, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", table)
	_, err := database.DB.Exec(query, id)
	if err != nil {
		logger.Error("hard delete failed", "table", table, "id", id, "error", err)
	}
	return err
}

func (b *Base) Exists(table string, where string, args []interface{}) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s)", table, where)
	var exists bool
	if err := database.DB.QueryRow(query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (b *Base) InClause(placeholder_start int, items []string) (string, []interface{}) {
	var placeholders []string
	var args []interface{}
	for i, item := range items {
		placeholders = append(placeholders, fmt.Sprintf("$%d", placeholder_start+i))
		args = append(args, item)
	}
	return strings.Join(placeholders, ", "), args
}
