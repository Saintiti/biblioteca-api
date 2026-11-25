package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"biblioteca-api/internal/models"
	"biblioteca-api/internal/repository"
	"biblioteca-api/internal/service"
	"github.com/gin-gonic/gin"
)

type LaptopHandler struct{ laptops *service.LaptopService }

func NewLaptopHandler(laptops *service.LaptopService) *LaptopHandler {
	return &LaptopHandler{laptops: laptops}
}

// Create registra un equipo (solo administradores).
func (h *LaptopHandler) Create(c *gin.Context) {
	var req models.LaptopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.laptops.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLaptopData) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, repository.ErrLaptopSerialExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo crear la laptop"})
		return
	}
	c.JSON(http.StatusCreated, item)
}
func (h *LaptopHandler) List(c *gin.Context) {
	items, err := h.laptops.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo consultar el inventario"})
		return
	}
	c.JSON(http.StatusOK, items)
}

type LaptopLoanHandler struct{ loans *service.LaptopLoanService }

func NewLaptopLoanHandler(loans *service.LaptopLoanService) *LaptopLoanHandler {
	return &LaptopLoanHandler{loans: loans}
}

func (h *LaptopLoanHandler) Borrow(c *gin.Context) {
	var req models.LaptopLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	loan, err := h.loans.Borrow(c.Request.Context(), c.GetInt("user_id"), req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrLaptopNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, repository.ErrLaptopUnavailable):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidLaptopData):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo prestar la laptop"})
		}
		return
	}
	c.JSON(http.StatusCreated, loan)
}
func (h *LaptopLoanHandler) MyLoans(c *gin.Context) {
	items, err := h.loans.MyLoans(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron listar los préstamos de laptops"})
		return
	}
	c.JSON(http.StatusOK, items)
}
func (h *LaptopLoanHandler) Return(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	err = h.loans.Return(c.Request.Context(), id, c.GetInt("user_id"), c.GetString("role"))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrLaptopLoanNotFound), errors.Is(err, sql.ErrNoRows):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrLaptopForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, repository.ErrLaptopAlreadyReturned):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo devolver la laptop"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "laptop devuelta correctamente"})
}
