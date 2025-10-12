package history

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"HR-Kafka-QA/internal/dto"
)

type PgxPoolIface interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	pool PgxPoolIface
}

func NewRepository(pool PgxPoolIface) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Insert(ctx context.Context, h dto.EmploymentHistory) error {
	q := `
INSERT INTO employment_history
	(employee_id, company, position, period_from, period_to, stack, created_at)
VALUES
	($1, $2, $3, $4::date, $5::date, $6, NOW());
`
	_, err := r.pool.Exec(ctx, q, h.EmployeeID, h.Company, h.Position, h.PeriodFrom, h.PeriodTo, h.Stack)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, h dto.EmploymentHistory) error {
	q := `
UPDATE employment_history SET
	employee_id = $2,
	company = $3,
	position = $4,
	period_from = $5::date,
	period_to = $6::date,
	stack = $7
WHERE id = $1`
	_, err := r.pool.Exec(ctx, q,
		h.ID, h.EmployeeID, h.Company, h.Position, h.PeriodFrom, h.PeriodTo, h.Stack,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	q := `DELETE FROM employment_history WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) ListByEmployee(ctx context.Context, employeeID string, limit, offset int) ([]dto.EmploymentHistory, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	q := `
SELECT id,
	   employee_id,
	   company,
	   position,
	   to_char(period_from,'YYYY-MM-DD'),
	   to_char(period_to,'YYYY-MM-DD'),
	   stack,
	   to_char(created_at,'YYYY-MM-DD"T"HH24:MI:SSOF')
FROM employment_history
WHERE employee_id = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3
`
	rows, err := r.pool.Query(ctx, q, employeeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dto.EmploymentHistory
	for rows.Next() {
		var it dto.EmploymentHistory

		err = rows.Scan(&it.ID, &it.EmployeeID, &it.Company, &it.Position, &it.PeriodFrom, &it.PeriodTo, &it.Stack, &it.CreatedAt)
		if err != nil {
			return nil, err
		}

		out = append(out, it)
	}

	return out, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*dto.EmploymentHistory, error) {
	q := `
SELECT id,
	   employee_id,
	   company,
	   position,
	   to_char(period_from,'YYYY-MM-DD'),
	   to_char(period_to,'YYYY-MM-DD'),
	   stack,
	   to_char(created_at,'YYYY-MM-DD"T"HH24:MI:SSOF')
FROM employment_history
WHERE id = $1;
`
	row := r.pool.QueryRow(ctx, q, id)

	var it dto.EmploymentHistory
	err := row.Scan(&it.ID, &it.EmployeeID, &it.Company, &it.Position, &it.PeriodFrom, &it.PeriodTo, &it.Stack, &it.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &it, nil
}
