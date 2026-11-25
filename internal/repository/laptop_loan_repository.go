package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"biblioteca-api/internal/models"
	"github.com/jmoiron/sqlx"
)

var (
	ErrLaptopUnavailable     = errors.New("la laptop no está disponible")
	ErrLaptopNotFound        = errors.New("laptop no encontrada")
	ErrLaptopLoanNotFound    = errors.New("préstamo de laptop no encontrado")
	ErrLaptopAlreadyReturned = errors.New("este préstamo ya fue devuelto")
)

type LaptopLoanRepository struct{ db *sqlx.DB }

func NewLaptopLoanRepository(db *sqlx.DB) *LaptopLoanRepository { return &LaptopLoanRepository{db: db} }

func (r *LaptopLoanRepository) FindByID(ctx context.Context, id int) (*models.LaptopLoan, error) {
	var loan models.LaptopLoan
	err := r.db.GetContext(ctx, &loan, "SELECT id,user_id,laptop_id,loaned_at,due_at,returned_at,status FROM laptop_loans WHERE id=?", id)
	return &loan, err
}

// BorrowAtomic contiene ambas escrituras en una transacción InnoDB.
// FOR UPDATE serializa dos peticiones concurrentes sobre la misma laptop.
func (r *LaptopLoanRepository) BorrowAtomic(ctx context.Context, userID, laptopID int, dueAt time.Time) (*models.LaptopLoan, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var status string
	err = tx.GetContext(ctx, &status, "SELECT status FROM laptops WHERE id=? FOR UPDATE", laptopID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrLaptopNotFound
	}
	if err != nil {
		return nil, err
	}
	if status != "available" {
		return nil, ErrLaptopUnavailable
	}
	res, err := tx.ExecContext(ctx, "UPDATE laptops SET status='borrowed' WHERE id=? AND status='available'", laptopID)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected != 1 {
		return nil, ErrLaptopUnavailable
	}
	res, err = tx.ExecContext(ctx,
		"INSERT INTO laptop_loans (user_id,laptop_id,due_at,status) VALUES (?,?,?,'active')",
		userID, laptopID, dueAt)
	if err != nil {
		return nil, err
	}
	loanID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, int(loanID))
}

// ReturnAtomic devuelve el equipo y cierra el préstamo en la misma transacción.
func (r *LaptopLoanRepository) ReturnAtomic(ctx context.Context, id int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var item struct {
		LaptopID int    `db:"laptop_id"`
		Status   string `db:"status"`
	}
	err = tx.GetContext(ctx, &item, "SELECT laptop_id,status FROM laptop_loans WHERE id=? FOR UPDATE", id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrLaptopLoanNotFound
	}
	if err != nil {
		return err
	}
	if item.Status == "returned" {
		return ErrLaptopAlreadyReturned
	}
	_, err = tx.ExecContext(ctx, "UPDATE laptop_loans SET status='returned', returned_at=NOW() WHERE id=? AND status='active'", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE laptops SET status='available' WHERE id=?", item.LaptopID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *LaptopLoanRepository) FindByUser(ctx context.Context, userID int) ([]models.LaptopLoanDetail, error) {
	items := make([]models.LaptopLoanDetail, 0)
	query := `SELECT ll.id,ll.user_id,ll.laptop_id,ll.loaned_at,ll.due_at,ll.returned_at,ll.status,
	               l.serial_number,l.brand,l.model
	        FROM laptop_loans ll JOIN laptops l ON l.id=ll.laptop_id
	        WHERE ll.user_id=? ORDER BY ll.loaned_at DESC`
	err := r.db.SelectContext(ctx, &items, query, userID)
	return items, err
}
