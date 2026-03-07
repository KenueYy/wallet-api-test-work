package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KenueYy/wallet-api/internal/config"
	"github.com/KenueYy/wallet-api/internal/db"
	"github.com/KenueYy/wallet-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	createOperationRoute = "/api/v1/wallet"
	getWalletRoute       = "/api/v1/wallets/"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterRoutes(r)
	initTestDB(t)
	return r
}

func initTestDB(t *testing.T) {
	t.Helper()

	cfg := &config.Config{
		DBHost:     "postgres_test",
		DBPort:     5432,
		DBUser:     "postgres",
		DBPassword: "password",
		DBName:     "walletdb_test",
		DBSSLMode:  "disable",
	}

	if err := db.Init(cfg); err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}

	if err := db.DB.Exec("TRUNCATE TABLE wallets RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to truncate wallets: %v", err)
	}
}

func TestCreateOperations_NegativeAmount(t *testing.T) {
	r := setupRouter(t)

	walletID := uuid.New()

	err := db.DB.Create(&models.Wallet{
		ID:      walletID,
		Balance: 1000,
	}).Error
	require.NoError(t, err)

	body := models.Operation{
		WalletID:  walletID,
		Operation: models.DEPOSIT,
		Amount:    -100,
	}

	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, createOperationRoute, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOperation_DepositWallet(t *testing.T) {
	r := setupRouter(t)

	walletID := uuid.New()

	err := db.DB.Create(&models.Wallet{
		ID:      walletID,
		Balance: 1000,
	}).Error
	require.NoError(t, err)

	body := models.Operation{
		WalletID:  walletID,
		Operation: models.DEPOSIT,
		Amount:    500,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, createOperationRoute, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var wallet models.Wallet
	err = db.DB.First(&wallet, "id = ?", walletID).Error
	require.NoError(t, err)
	require.Equal(t, int64(1500), wallet.Balance)
}

func TestCreateOperation_WithdrawNotFound(t *testing.T) {
	r := setupRouter(t)

	body := models.Operation{
		WalletID:  uuid.New(),
		Operation: models.WITHDRAW,
		Amount:    100,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, createOperationRoute, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetWallet_OK(t *testing.T) {
	r := setupRouter(t)

	walletID := uuid.New()
	err := db.DB.Create(&models.Wallet{
		ID:      walletID,
		Balance: 2000,
	}).Error
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, getWalletRoute+walletID.String(), nil)

	r.ServeHTTP(w, req)

	var resp models.Wallet
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, walletID, resp.ID)
	require.Equal(t, int64(2000), resp.Balance)
}

func TestCreateOperation_WithdrawInsufficientFunds(t *testing.T) {
	r := setupRouter(t)

	walletID := uuid.New()

	err := db.DB.Create(&models.Wallet{
		ID:      walletID,
		Balance: 0,
	}).Error
	require.NoError(t, err)

	body := models.Operation{
		WalletID:  walletID,
		Operation: models.WITHDRAW,
		Amount:    500,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, createOperationRoute, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var wallet models.Wallet
	err = db.DB.First(&wallet, "id=?", walletID).Error
	require.NoError(t, err)
	require.Equal(t, int64(0), wallet.Balance)
}

func TestGetWallet_InvalidID(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, getWalletRoute+"123-321-123-000", nil)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOperation_InvalidJSON(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	body := []byte(`{`)

	req, _ := http.NewRequest(http.MethodPost, createOperationRoute, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOperation_MissingFields(t *testing.T) {
	r := setupRouter(t)

	payload := map[string]any{
		"operationType": "DEPOSIT",
		"amount":        0,
	}
	b, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, createOperationRoute, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
