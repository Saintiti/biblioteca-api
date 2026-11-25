package repository

import (
	"context"
	"database/sql"
	"errors"

	"biblioteca-api/internal/models"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var ErrLaptopSerialExists = errors.New("ya existe una laptop con ese serial")

type LaptopRepository struct{ db *sqlx.DB }

func NewLaptopRepository(db *sqlx.DB) *LaptopRepository { return &LaptopRepository{db: db} }

func (r *LaptopRepository) Create(ctx context.Context, req models.LaptopRequest) (*models.Laptop, error) {
	var existing int
	err := r.db.GetContext(ctx, &existing, "SELECT id FROM laptops WHERE serial_number = ?", req.SerialNumber)
	if err == nil {
		return nil, ErrLaptopSerialExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO laptops (serial_number, brand, model) VALUES (?,?,?)",
		req.SerialNumber, req.Brand, req.Model)
	if err != nil {
		var mysqlErr *mysqldriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrLaptopSerialExists
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, int(id))
}

func (r *LaptopRepository) FindByID(ctx context.Context, id int) (*models.Laptop, error) {
	var laptop models.Laptop
	if err := r.db.GetContext(ctx, &laptop, "SELECT * FROM laptops WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &laptop, nil
}

func (r *LaptopRepository) List(ctx context.Context) ([]models.Laptop, error) {
	items := make([]models.Laptop, 0)
	err := r.db.SelectContext(ctx, &items, "SELECT * FROM laptops ORDER BY id DESC")
	return items, err
}
