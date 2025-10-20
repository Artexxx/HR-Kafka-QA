package history

import (
	"context"
	"errors"
	"fmt"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (r *Repository) Insert(ctx context.Context, history dto.EmploymentHistory) error {
	query := `
insert into employment_history
  (employee_id, company, position, period_from, period_to, stack, created_at)
values
  (@employee_id, @company, @position, @period_from::date, @period_to::date, @stack, now());
`
	args := pgx.NamedArgs{
		"employee_id": history.EmployeeID,
		"company":     history.Company,
		"position":    history.Position,
		"period_from": history.PeriodFrom,
		"period_to":   history.PeriodTo,
		"stack":       history.Stack,
	}

	_, err := r.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, history dto.EmploymentHistory) error {
	query := `
update employment_history set
  employee_id = @employee_id,
  company     = @company,
  position    = @position,
  period_from = @period_from::date,
  period_to   = @period_to::date,
  stack       = @stack
where id = @id;
`
	args := pgx.NamedArgs{
		"id":          history.ID,
		"employee_id": history.EmployeeID,
		"company":     history.Company,
		"position":    history.Position,
		"period_from": history.PeriodFrom,
		"period_to":   history.PeriodTo,
		"stack":       history.Stack,
	}

	tag, err := r.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return dto.ErrNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	query := `delete from employment_history where id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return dto.ErrNotFound
	}

	return nil
}

func (r *Repository) ListByEmployee(ctx context.Context, employeeID string) ([]dto.EmploymentHistory, error) {
	query := `
select id,
	   employee_id,
	   company,
	   position,
	   to_char(period_from,'YYYY-MM-DD'),
	   to_char(period_to,'YYYY-MM-DD'),
	   stack,
	   created_at
from employment_history
where employee_id = $1
order by id desc
`
	rows, err := r.pool.Query(ctx, query, employeeID)
	if err != nil {
		return nil, fmt.Errorf("pool.Query: %w", err)
	}
	defer rows.Close()

	var out []dto.EmploymentHistory
	for rows.Next() {
		var history dto.EmploymentHistory

		err = rows.Scan(&history.ID, &history.EmployeeID, &history.Company, &history.Position, &history.PeriodFrom, &history.PeriodTo, &history.Stack, &history.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		out = append(out, history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*dto.EmploymentHistory, error) {
	query := `
select id,
	   employee_id,
	   company,
	   position,
	   to_char(period_from,'YYYY-MM-DD'),
	   to_char(period_to,'YYYY-MM-DD'),
	   stack,
	   created_at
from employment_history
where id = $1;
`
	row := r.pool.QueryRow(ctx, query, id)

	var history dto.EmploymentHistory
	err := row.Scan(&history.ID, &history.EmployeeID, &history.Company, &history.Position, &history.PeriodFrom, &history.PeriodTo, &history.Stack, &history.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dto.ErrNotFound
		}

		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return &history, nil
}
