package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/KenueYy/wallet-api/internal/db"
	"github.com/KenueYy/wallet-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInvalidOperation  = errors.New("invalid operations")
)

func RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	v1.POST("/wallet", createOperation)
	v1.GET("/wallets/:id", getWallet)
}

// post
func createOperation(c *gin.Context) {
	var request models.Operation

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be positive"})
		return
	}

	var wallet models.Wallet
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&wallet, "id=?", request.WalletID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if request.Operation == models.WITHDRAW {
					return ErrWalletNotFound
				}

				wallet.ID = request.WalletID
				wallet.Balance = 0
				if err := tx.Create(&wallet).Error; err != nil { // в проде бы так не делал, использовал бы только явное создание кошелька
					return err
				}
			}
		}

		switch request.Operation {
		case models.DEPOSIT:
			wallet.Balance += request.Amount
		case models.WITHDRAW:
			if wallet.Balance < request.Amount {
				return ErrInsufficientFunds
			}
			wallet.Balance -= request.Amount
		default:
			return ErrInvalidOperation
		}
		return tx.Save(&wallet).Error
	}); err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrInsufficientFunds.Error()})
			return
		}

		if errors.Is(err, ErrInvalidOperation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidOperation.Error()})
			return
		}

		if errors.Is(err, ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrWalletNotFound.Error()})
			return
		}

		log.Println("Create wallet error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"walletId":  request.WalletID,
		"balance":   wallet.Balance,
		"operation": request.Operation,
		"amount":    request.Amount,
	})
}

// get
func getWallet(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var wallet models.Wallet
	if err := db.DB.First(&wallet, "id=?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrWalletNotFound.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      wallet.ID,
		"balance": wallet.Balance,
	})
}
