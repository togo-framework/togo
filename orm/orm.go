// Package orm is a small, driver-agnostic, Eloquent-style query builder over
// database/sql. The dialect (placeholder style, ILIKE handling) is chosen from
// the configured driver, so models are written once and the driver is swapped
// from .env (DB_DRIVER) without changing code.
//
//	models.Posts(app).Find(ctx, id)
//	models.Posts(app).Where("title", "ILIKE", "%go%").Order("created_at DESC").Get(ctx)
package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Dialect captures the per-driver SQL differences the builder needs.
type Dialect struct {
	Placeholder func(n int) string // 1-based positional placeholder
	ILike       string             // case-insensitive LIKE operator
}

var (
	sqliteDialect   = Dialect{Placeholder: func(int) string { return "?" }, ILike: "LIKE"}
	postgresDialect = Dialect{Placeholder: func(n int) string { return "$" + strconv.Itoa(n) }, ILike: "ILIKE"}
	mysqlDialect    = Dialect{Placeholder: func(int) string { return "?" }, ILike: "LIKE"}
)

// DialectFor returns the dialect for a database/sql driver name.
func DialectFor(driver string) Dialect {
	switch driver {
	case "pgx", "postgres", "postgresql":
		return postgresDialect
	case "mysql":
		return mysqlDialect
	default:
		return sqliteDialect
	}
}

type cond struct {
	col, op string
}

// Query is a fluent, typed query builder for table rows scanned into T.
type Query[T any] struct {
	db     *sql.DB
	d      Dialect
	table  string
	conds  []cond
	args   []any
	order  string
	limit  int
	offset int
}

// For starts a query against table, scanning into T.
func For[T any](db *sql.DB, d Dialect, table string) *Query[T] {
	return &Query[T]{db: db, d: d, table: table, limit: -1, offset: -1}
}

// Where adds a condition. op may be =, !=, <, >, LIKE, ILIKE, etc.
func (q *Query[T]) Where(col, op string, val any) *Query[T] {
	if strings.EqualFold(op, "ILIKE") {
		op = q.d.ILike
	}
	q.conds = append(q.conds, cond{col, op})
	q.args = append(q.args, val)
	return q
}

// Order sets the ORDER BY clause (raw, e.g. "created_at DESC").
func (q *Query[T]) Order(s string) *Query[T] { q.order = s; return q }

// Limit sets LIMIT.
func (q *Query[T]) Limit(n int) *Query[T] { q.limit = n; return q }

// Offset sets OFFSET.
func (q *Query[T]) Offset(n int) *Query[T] { q.offset = n; return q }

func (q *Query[T]) where() (string, []any) {
	if len(q.conds) == 0 {
		return "", nil
	}
	parts := make([]string, len(q.conds))
	for i, c := range q.conds {
		parts[i] = fmt.Sprintf("%s %s %s", c.col, c.op, q.d.Placeholder(i+1))
	}
	return " WHERE " + strings.Join(parts, " AND "), q.args
}

// Get returns all matching rows.
func (q *Query[T]) Get(ctx context.Context) ([]T, error) {
	w, args := q.where()
	sb := "SELECT * FROM " + q.table + w
	if q.order != "" {
		sb += " ORDER BY " + q.order
	}
	if q.limit >= 0 {
		sb += " LIMIT " + strconv.Itoa(q.limit)
	}
	if q.offset >= 0 {
		sb += " OFFSET " + strconv.Itoa(q.offset)
	}
	rows, err := q.db.QueryContext(ctx, sb, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAll[T](rows)
}

// First returns the first matching row, or (nil, nil) if none.
func (q *Query[T]) First(ctx context.Context) (*T, error) {
	q.limit = 1
	rows, err := q.Get(ctx)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0], nil
}

// Find returns the row with the given id.
func (q *Query[T]) Find(ctx context.Context, id any) (*T, error) {
	return q.Where("id", "=", id).First(ctx)
}

// Create inserts a row and returns it (RETURNING *).
func (q *Query[T]) Create(ctx context.Context, data map[string]any) (*T, error) {
	cols := make([]string, 0, len(data))
	ph := make([]string, 0, len(data))
	args := make([]any, 0, len(data))
	i := 1
	for c, v := range data {
		cols = append(cols, c)
		ph = append(ph, q.d.Placeholder(i))
		args = append(args, v)
		i++
	}
	sb := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING *", q.table, strings.Join(cols, ", "), strings.Join(ph, ", "))
	rows, err := q.db.QueryContext(ctx, sb, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out, err := scanAll[T](rows)
	if err != nil || len(out) == 0 {
		return nil, err
	}
	return &out[0], nil
}

// Delete deletes matching rows.
func (q *Query[T]) Delete(ctx context.Context) error {
	w, args := q.where()
	_, err := q.db.ExecContext(ctx, "DELETE FROM "+q.table+w, args...)
	return err
}

// scanAll scans all rows into []T by matching columns to `db` struct tags.
func scanAll[T any](rows *sql.Rows) ([]T, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var out []T
	for rows.Next() {
		var item T
		v := reflect.ValueOf(&item).Elem()
		fieldByCol := dbFields(v.Type())
		dest := make([]any, len(cols))
		var discard any
		for i, c := range cols {
			if idx, ok := fieldByCol[c]; ok {
				dest[i] = v.Field(idx).Addr().Interface()
			} else {
				dest[i] = &discard
			}
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// dbFields maps column name → struct field index via the `db` tag.
func dbFields(t reflect.Type) map[string]int {
	m := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("db"); tag != "" && tag != "-" {
			m[tag] = i
		}
	}
	return m
}
