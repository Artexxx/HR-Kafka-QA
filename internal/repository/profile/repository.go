package profile

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
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

func (r *Repository) Create(ctx context.Context, p dto.EmployeeProfile) error {
	query := `
insert into employee_profile
  (employee_id, first_name, last_name, birth_date, email, phone, title, department, grade, effective_from, updated_at)
values
  (@employee_id, @first_name, @last_name, nullif(@birth_date, '')::date, @email, @phone, @title, @department, @grade, nullif(@effective_from, '')::date, now());
`
	args := pgx.NamedArgs{
		"employee_id":    p.EmployeeID,
		"first_name":     p.FirstName,
		"last_name":      p.LastName,
		"birth_date":     strptr(p.BirthDate),
		"email":          p.Email,
		"phone":          p.Phone,
		"title":          p.Title,
		"department":     p.Department,
		"grade":          p.Grade,
		"effective_from": strptr(p.EffectiveFrom),
	}

	_, err := r.pool.Exec(ctx, query, args)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23505" {
			return dto.ErrAlreadyExists
		}

		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, p dto.EmployeeProfile) error {
	query := `
update employee_profile set
  first_name     = @first_name,
  last_name      = @last_name,
  birth_date     = nullif(@birth_date,'')::date,
  email          = @email,
  phone          = @phone,
  title          = @title,
  department     = @department,
  grade          = @grade,
  effective_from = nullif(@effective_from,'')::date,
  updated_at     = now()
where employee_id = @employee_id;
`
	args := pgx.NamedArgs{
		"employee_id":    p.EmployeeID,
		"first_name":     p.FirstName,
		"last_name":      p.LastName,
		"birth_date":     strptr(p.BirthDate),
		"email":          p.Email,
		"phone":          p.Phone,
		"title":          p.Title,
		"department":     p.Department,
		"grade":          p.Grade,
		"effective_from": strptr(p.EffectiveFrom),
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

func (r *Repository) Delete(ctx context.Context, employeeID string) error {
	query := `delete from employee_profile where employee_id = $1`

	tag, err := r.pool.Exec(ctx, query, employeeID)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return dto.ErrNotFound
	}

	return nil
}

func (r *Repository) GetProfile(ctx context.Context, employeeID string) (*dto.EmployeeProfile, error) {
	query := `
select employee_id,
	   first_name,
	   last_name,
	   to_char(birth_date,'YYYY-MM-DD'),
	   email,
	   phone,
	   title,
	   department,
	   grade,
	   to_char(effective_from,'YYYY-MM-DD'),
	   to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
from employee_profile
where employee_id = $1;
`
	row := r.pool.QueryRow(ctx, query, employeeID)

	var (
		out           dto.EmployeeProfile
		birthDate     *string
		effectiveFrom *string
	)

	err := row.Scan(
		&out.EmployeeID,
		&out.FirstName,
		&out.LastName,
		&birthDate,
		&out.Email,
		&out.Phone,
		&out.Title,
		&out.Department,
		&out.Grade,
		&effectiveFrom,
		&out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dto.ErrNotFound
		}

		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	out.BirthDate = birthDate
	out.EffectiveFrom = effectiveFrom

	return &out, nil
}

func (r *Repository) ListProfiles(ctx context.Context) ([]dto.EmployeeProfile, error) {
	query := `
select employee_id,
       first_name,
       last_name,
       to_char(birth_date,'YYYY-MM-DD'),
       email,
       phone,
       title,
       department,
       grade,
       to_char(effective_from,'YYYY-MM-DD'),
       to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
from employee_profile
order by updated_at desc, employee_id
`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("pool.Query: %w", err)
	}
	defer rows.Close()

	var out []dto.EmployeeProfile
	for rows.Next() {
		var (
			profile                  dto.EmployeeProfile
			birthDate, effectiveFrom *string
		)

		err = rows.Scan(
			&profile.EmployeeID,
			&profile.FirstName,
			&profile.LastName,
			&birthDate,
			&profile.Email,
			&profile.Phone,
			&profile.Title,
			&profile.Department,
			&profile.Grade,
			&effectiveFrom,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		profile.BirthDate = birthDate
		profile.EffectiveFrom = effectiveFrom
		out = append(out, profile)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return out, nil
}

func (r *Repository) UpsertPersonal(ctx context.Context, p dto.EmployeeProfile) error {
	query := `
insert into employee_profile (employee_id, first_name, last_name, birth_date, email, phone, updated_at)
values (@employee_id, @first_name, @last_name, nullif(@birth_date,'')::date, @email, @phone, now())
on conflict (employee_id) do update set
  first_name = excluded.first_name,
  last_name  = excluded.last_name,
  birth_date = excluded.birth_date,
  email      = excluded.email,
  phone      = excluded.phone,
  updated_at = now();
`
	args := pgx.NamedArgs{
		"employee_id": p.EmployeeID,
		"first_name":  p.FirstName,
		"last_name":   p.LastName,
		"birth_date":  strptr(p.BirthDate),
		"email":       p.Email,
		"phone":       p.Phone,
	}

	if _, err := r.pool.Exec(ctx, query, args); err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func (r *Repository) UpsertPosition(ctx context.Context, p dto.EmployeeProfile) error {
	query := `
insert into employee_profile (employee_id, title, department, grade, effective_from, updated_at)
values (@employee_id, @title, @department, @grade, nullif(@effective_from,'')::date, now())
on conflict (employee_id) do update set
  title          = excluded.title,
  department     = excluded.department,
  grade          = excluded.grade,
  effective_from = excluded.effective_from,
  updated_at     = now();
`
	args := pgx.NamedArgs{
		"employee_id":    p.EmployeeID,
		"title":          p.Title,
		"department":     p.Department,
		"grade":          p.Grade,
		"effective_from": strptr(p.EffectiveFrom),
	}

	if _, err := r.pool.Exec(ctx, query, args); err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func strptr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
