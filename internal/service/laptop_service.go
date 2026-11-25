package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"biblioteca-api/internal/models"
	"biblioteca-api/internal/repository"
)

var (
	ErrInvalidLaptopData = errors.New("serial, marca y modelo son obligatorios")
	ErrLaptopForbidden   = errors.New("solo el dueño del préstamo o un administrador puede devolver la laptop")
)

const laptopLoanDurationDays = 7

type LaptopService struct{ repo *repository.LaptopRepository }

func NewLaptopService(repo *repository.LaptopRepository) *LaptopService {
	return &LaptopService{repo: repo}
}
func (s *LaptopService) Create(ctx context.Context, req models.LaptopRequest) (*models.Laptop, error) {
	req.SerialNumber = strings.TrimSpace(req.SerialNumber)
	req.Brand = strings.TrimSpace(req.Brand)
	req.Model = strings.TrimSpace(req.Model)
	if req.SerialNumber == "" || req.Brand == "" || req.Model == "" {
		return nil, ErrInvalidLaptopData
	}
	return s.repo.Create(ctx, req)
}
func (s *LaptopService) List(ctx context.Context) ([]models.Laptop, error) { return s.repo.List(ctx) }

type LaptopLoanService struct {
	laptops *repository.LaptopRepository
	loans   *repository.LaptopLoanRepository
}

func NewLaptopLoanService(laptops *repository.LaptopRepository, loans *repository.LaptopLoanRepository) *LaptopLoanService {
	return &LaptopLoanService{laptops: laptops, loans: loans}
}
func (s *LaptopLoanService) Borrow(ctx context.Context, userID int, req models.LaptopLoanRequest) (*models.LaptopLoan, error) {
	if req.LaptopID <= 0 {
		return nil, ErrInvalidLaptopData
	}
	laptop, err := s.laptops.FindByID(ctx, req.LaptopID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrLaptopNotFound
	}
	if err != nil {
		return nil, err
	}
	if laptop.Status != "available" {
		return nil, repository.ErrLaptopUnavailable
	}
	return s.loans.BorrowAtomic(ctx, userID, req.LaptopID, time.Now().AddDate(0, 0, laptopLoanDurationDays))
}
func (s *LaptopLoanService) Return(ctx context.Context, id, userID int, role string) error {
	loan, err := s.loans.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ErrLaptopLoanNotFound
	}
	if err != nil {
		return err
	}
	if loan.UserID != userID && role != "admin" {
		return ErrLaptopForbidden
	}
	if loan.Status == "returned" {
		return repository.ErrLaptopAlreadyReturned
	}
	return s.loans.ReturnAtomic(ctx, id)
}
func (s *LaptopLoanService) MyLoans(ctx context.Context, userID int) ([]models.LaptopLoanDetail, error) {
	return s.loans.FindByUser(ctx, userID)
}
