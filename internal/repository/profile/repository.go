package profile

import (
	"context"

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
	q := `
INSERT INTO employee_profile
	(employee_id, first_name, last_name, birth_date, email, phone, title, department, grade, effective_from, updated_at)
VALUES
	($1, $2, $3, NULLIF($4, '')::date, $5, $6, $7, $8, $9, NULLIF($10, '')::date, NOW());
`
	_, err := r.pool.Exec(ctx, q, p.EmployeeID, p.FirstName, p.LastName, strptr(p.BirthDate), p.Email, p.Phone, p.Title, p.Department, p.Grade, strptr(p.EffectiveFrom))
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, p dto.EmployeeProfile) error {
	q := `
UPDATE employee_profile SET
	first_name = COALESCE($2, first_name),
	last_name = COALESCE($3, last_name),
	birth_date = COALESCE(NULLIF($4,'')::date, birth_date),
	email = COALESCE($5, email),
	phone = COALESCE($6, phone),
	title = COALESCE($7, title),
	department = COALESCE($8, department),
	grade = COALESCE($9, grade),
	effective_from = COALESCE(NULLIF($10,'')::date, effective_from),
	updated_at = NOW()
WHERE employee_id = $1;
`
	_, err := r.pool.Exec(ctx, q,
		p.EmployeeID, p.FirstName, p.LastName, strptr(p.BirthDate),
		p.Email, p.Phone, p.Title, p.Department, p.Grade, strptr(p.EffectiveFrom),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, employeeID string) error {
	q := `DELETE FROM employee_profile WHERE employee_id = $1`
	_, err := r.pool.Exec(ctx, q, employeeID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetProfile(ctx context.Context, employeeID string) (*dto.EmployeeProfile, error) {
	q := `
SELECT employee_id,
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
FROM employee_profile
WHERE employee_id = $1;
`
	row := r.pool.QueryRow(ctx, q, employeeID)

	var out dto.EmployeeProfile
	var birthDate, effectiveFrom *string

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
		return nil, err
	}

	out.BirthDate = birthDate
	out.EffectiveFrom = effectiveFrom

	return &out, nil
}

func (r *Repository) ListProfiles(ctx context.Context, limit, offset int) ([]dto.EmployeeProfile, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	q := `
SELECT employee_id,
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
FROM employee_profile
ORDER BY updated_at DESC, employee_id
LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dto.EmployeeProfile
	for rows.Next() {
		var it dto.EmployeeProfile
		var birthDate, effectiveFrom *string
		err = rows.Scan(
			&it.EmployeeID,
			&it.FirstName,
			&it.LastName,
			&birthDate,
			&it.Email,
			&it.Phone,
			&it.Title,
			&it.Department,
			&it.Grade,
			&effectiveFrom,
			&it.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		it.BirthDate = birthDate
		it.EffectiveFrom = effectiveFrom
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertPersonal(ctx context.Context, p dto.EmployeeProfile) error {
	q := `
INSERT INTO employee_profile (employee_id, first_name, last_name, birth_date, email, phone, updated_at)
VALUES ($1, $2, $3, NULLIF($4,'')::date, $5, $6, NOW())
ON CONFLICT (employee_id) DO UPDATE SET
	first_name = COALESCE(EXCLUDED.first_name, employee_profile.first_name),
	last_name = COALESCE(EXCLUDED.last_name, employee_profile.last_name),
	birth_date = COALESCE(EXCLUDED.birth_date, employee_profile.birth_date),
	email = COALESCE(EXCLUDED.email, employee_profile.email),
	phone = COALESCE(EXCLUDED.phone, employee_profile.phone),
	updated_at = NOW();
`
	_, err := r.pool.Exec(ctx, q,
		p.EmployeeID, p.FirstName, p.LastName, strptr(p.BirthDate), p.Email, p.Phone,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpsertPosition(ctx context.Context, p dto.EmployeeProfile) error {
	q := `
INSERT INTO employee_profile (employee_id, title, department, grade, effective_from, updated_at)
VALUES ($1, $2, $3, $4, NULLIF($5,'')::date, NOW())
ON CONFLICT (employee_id) DO UPDATE SET
	title = COALESCE(EXCLUDED.title, employee_profile.title),
	department = COALESCE(EXCLUDED.department, employee_profile.department),
	grade = COALESCE(EXCLUDED.grade, employee_profile.grade),
	effective_from = COALESCE(EXCLUDED.effective_from, employee_profile.effective_from),
	updated_at = NOW();
`
	_, err := r.pool.Exec(ctx, q,
		p.EmployeeID, p.Title, p.Department, p.Grade, strptr(p.EffectiveFrom),
	)
	if err != nil {
		return err
	}

	return nil
}

func strptr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
