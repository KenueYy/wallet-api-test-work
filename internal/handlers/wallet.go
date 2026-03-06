package handlers

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"

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

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	v1.POST("/wallet", createOperation)
	v1.GET("/wallets/:id", getWallet)
}

// post
func createOperation(c *gin.Context) {
	var request models.Operation

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warn("invalid request",
			"wallet_id", request.WalletID.String(),
			"error", err.Error(),
			"path", c.FullPath(),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Amount <= 0 {
		logger.Warn("non-positive amount",
			"wallet_id", request.WalletID.String(),
			"amount", request.Amount,
			"path", c.FullPath(),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be positive"})
		return
	}

	var wallet models.Wallet
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&wallet, "id=?", request.WalletID).Error; err != nil {

			if err == gorm.ErrRecordNotFound {
				if request.Operation == models.WITHDRAW {
					logger.Warn("wallet not found on withdraw",
						"wallet_id", request.WalletID.String(),
						"path", c.FullPath(),
					)
					return ErrWalletNotFound
				}

				wallet.ID = request.WalletID
				wallet.Balance = 0

				if err := tx.Create(&wallet).Error; err != nil { // в проде бы так не делал, использовал бы только явное создание кошелька
					logger.Error("wallet create failed",
						"wallet_id", request.WalletID.String(),
						"error", err.Error(),
						"path", c.FullPath(),
					)
					return err
				}

				logger.Info("wallet auto-created",
					"wallet_id", request.WalletID.String(),
					"path", c.FullPath(),
				)
				return nil
			}

			logger.Error("transaction query failed",
				"wallet_id", request.WalletID.String(),
				"error", err.Error(),
				"path", c.FullPath(),
			)
			return err
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

		if err := tx.Save(&wallet).Error; err != nil {
			logger.Error("wallet save failed",
				"wallet_id", request.WalletID.String(),
				"error", err.Error(),
				"path", c.FullPath(),
			)
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			logger.Warn("insufficient funds",
				"wallet_id", request.WalletID.String(),
				"amount", request.Amount,
				"balance", wallet.Balance,
				"path", c.FullPath(),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrInsufficientFunds.Error()})
			return
		}

		if errors.Is(err, ErrInvalidOperation) {
			logger.Warn("invalid operation",
				"wallet_id", request.WalletID.String(),
				"operation", string(request.Operation),
				"path", c.FullPath(),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidOperation.Error()})
			return
		}

		if errors.Is(err, ErrWalletNotFound) {
			logger.Warn("wallet not found",
				"wallet_id", request.WalletID.String(),
				"path", c.FullPath(),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": ErrWalletNotFound.Error()})
			return
		}

		logger.Error("create operation failed",
			"wallet_id", request.WalletID.String(),
			"operation", string(request.Operation),
			"amount", request.Amount,
			"error", err.Error(),
			"path", c.FullPath(),
		)

		log.Println("Create wallet error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	logger.Info("operation applied",
		"wallet_id", request.WalletID.String(),
		"operation", string(request.Operation),
		"amount", request.Amount,
		"new_balance", wallet.Balance,
		"path", c.FullPath(),
	)

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
		logger.Warn("invalid wallet id",
			"id_raw", idParam,
			"path", c.FullPath(),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var wallet models.Wallet
	if err := db.DB.First(&wallet, "id=?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("wallet not found",
				"wallet_id", id.String(),
				"path", c.FullPath(),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": ErrWalletNotFound.Error()})
			return
		}

		logger.Error("get wallet failed",
			"wallet_id", id.String(),
			"error", err.Error(),
			"path", c.FullPath(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	logger.Info("wallet fetched",
		"wallet_id", wallet.ID.String(),
		"balance", wallet.Balance,
		"path", c.FullPath(),
	)

	c.JSON(http.StatusOK, gin.H{
		"id":      wallet.ID,
		"balance": wallet.Balance,
	})
}
