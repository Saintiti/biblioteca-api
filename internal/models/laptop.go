package models

import "time"

// Laptop es un equipo físico e individual, identificado por su serial.
type Laptop struct {
	ID           int       `db:"id" json:"id"`
	SerialNumber string    `db:"serial_number" json:"serial_number"`
	Brand        string    `db:"brand" json:"brand"`
	Model        string    `db:"model" json:"model"`
	Status       string    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// LaptopRequest representa los datos que puede registrar un administrador.
type LaptopRequest struct {
	SerialNumber string `json:"serial_number" binding:"required"`
	Brand        string `json:"brand" binding:"required"`
	Model        string `json:"model" binding:"required"`
}

// LaptopLoanRequest solicita prestar un equipo existente.
type LaptopLoanRequest struct {
	LaptopID int `json:"laptop_id" binding:"required,min=1"`
}

// LaptopLoan almacena fechas y estado del préstamo individual.
type LaptopLoan struct {
	ID         int        `db:"id" json:"id"`
	UserID     int        `db:"user_id" json:"user_id"`
	LaptopID   int        `db:"laptop_id" json:"laptop_id"`
	LoanedAt   time.Time  `db:"loaned_at" json:"loaned_at"`
	DueAt      time.Time  `db:"due_at" json:"due_at"`
	ReturnedAt *time.Time `db:"returned_at" json:"returned_at,omitempty"`
	Status     string     `db:"status" json:"status"`
}

// LaptopLoanDetail suma el serial/modelo para que el cliente no haga otra consulta.
type LaptopLoanDetail struct {
	LaptopLoan
	SerialNumber string `db:"serial_number" json:"serial_number"`
	Brand        string `db:"brand" json:"brand"`
	Model        string `db:"model" json:"model"`
}
