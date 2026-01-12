package wish

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Postgres
type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) Create(ctx context.Context, wish *Wish) error {
	const query = `
		INSERT INTO wishes (owner_email, title, description, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		wish.OwnerEmail,
		wish.Title,
		wish.Description,
		wish.CreatedAt,
	).Scan(&wish.ID)

	return err
}

func (r *PostgresRepo) GetByID(ctx context.Context, id int) (*Wish, error) {
	const query = `
		SELECT id, owner_email, title, description, is_bought, bought_at, created_at
		FROM wishes
		WHERE id = $1
	`

	wish := &Wish{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wish.ID,
		&wish.OwnerEmail,
		&wish.Title,
		&wish.Description,
		&wish.IsBought,
		&wish.BoughtAt,
		&wish.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrWishNotFound
	}

	return wish, err
}

func (r *PostgresRepo) List(ctx context.Context, ownerEmail string, bought *bool) ([]*Wish, error) {
	query := `
		SELECT id, owner_email, title, description, is_bought, bought_at, created_at
		FROM wishes
		WHERE owner_email = $1
	`
	args := []interface{}{ownerEmail}
	argPos := 2

	if bought != nil {
		query += fmt.Sprintf(" AND is_bought = $%d", argPos)
		args = append(args, *bought)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wishes []*Wish
	for rows.Next() {
		wish := &Wish{}
		err := rows.Scan(
			&wish.ID,
			&wish.OwnerEmail,
			&wish.Title,
			&wish.Description,
			&wish.IsBought,
			&wish.BoughtAt,
			&wish.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		wishes = append(wishes, wish)
	}

	return wishes, rows.Err()
}

func (r *PostgresRepo) Update(ctx context.Context, wish *Wish) error {
	const query = `
		UPDATE wishes
		set title = $1, description = $2, is_bought = $3, bought_at = $4
		WHERE id = $5
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		wish.Title,
		wish.Description,
		wish.IsBought,
		wish.BoughtAt,
		wish.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrWishNotFound
	}

	return nil
}

func (r *PostgresRepo) Delete(ctx context.Context, id int) error {
	const query = `
		DELETE FROM wishes
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrWishNotFound
	}

	return nil
}

func (r *PostgresRepo) Stats(ctx context.Context, ownerEmail string) (Stats, error) {
	const query = `
		SELECT 
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE is_bought = true) AS bought
		FROM wishes
		WHERE owner_email = $1
	`
	var total, bought int
	err := r.db.QueryRowContext(ctx, query, ownerEmail).Scan(&total, &bought)
	if err != nil {
		return Stats{}, err
	}

	return Stats{
		Total:   total,
		Bought:  bought,
		Pending: total - bought,
	}, nil
}
