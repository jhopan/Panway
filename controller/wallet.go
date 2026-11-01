package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/ctxkey"
	"github.com/jhopan/Panway/common/walletclient"
	"github.com/jhopan/Panway/model"
)

// walletClient lazily builds the PanWallet client from config. It returns nil
// when PanWallet integration is disabled, so every caller must handle that.
func walletClient() *walletclient.Client {
	if !config.PanWalletEnabled || config.PanWalletURL == "" {
		return nil
	}
	return walletclient.New(config.PanWalletURL, config.PanWalletKey)
}

// LinkWalletRequest is the payload for linking a PanWallet wallet.
type LinkWalletRequest struct {
	PublicID string `json:"public_id"`
}

// LinkWallet links the logged-in user to a PanWallet wallet.
//
// The customer enters their short PanWallet id. Panway verifies it against the
// PanWallet API before storing, so a typo cannot silently link the wrong
// wallet.
func LinkWallet(c *gin.Context) {
	if walletClient() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Integrasi PanWallet belum diaktifkan",
		})
		return
	}

	var req LinkWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	publicID := strings.ToUpper(strings.TrimSpace(req.PublicID))
	if publicID == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "ID wallet wajib diisi"})
		return
	}

	bal, err := walletClient().GetBalance(publicID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Wallet tidak ditemukan: " + err.Error(),
		})
		return
	}
	if bal.Status != "" && bal.Status != "active" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Wallet diblokir"})
		return
	}

	userId := c.GetInt(ctxkey.Id)
	if existing, err := model.GetWalletLinkByUserId(userId); err == nil && existing.PublicID != publicID {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Akun ini sudah tertaut ke wallet lain. Hubungi admin untuk melepas tautan.",
		})
		return
	}

	if err := model.UpsertWalletLink(userId, publicID, bal.PublicID); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"public_id": publicID,
			"balance":   bal.Balance,
			"available": bal.Available,
			"currency":  bal.Currency,
		},
	})
}

// UnlinkWallet removes the wallet link of the logged-in user.
func UnlinkWallet(c *gin.Context) {
	if err := model.DeleteWalletLink(c.GetInt(ctxkey.Id)); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// GetWallet returns the linked wallet and its live balance.
func GetWallet(c *gin.Context) {
	link, err := model.GetWalletLinkByUserId(c.GetInt(ctxkey.Id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Wallet belum ditautkan",
		})
		return
	}
	data := gin.H{
		"public_id":     link.PublicID,
		"wallet_user_id": link.WalletUserID,
	}
	if wc := walletClient(); wc != nil {
		if bal, err := wc.GetBalance(link.PublicID); err == nil {
			data["balance"] = bal.Balance
			data["available"] = bal.Available
			data["held"] = bal.Held
			data["currency"] = bal.Currency
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": data})
}

// SyncWalletTopUp mirrors a paid amount from PanWallet into Panway quota.
//
// The customer pays a QRIS invoice in PanWallet, then redeems the amount
// here. Panway verifies the balance with PanWallet before crediting, so a
// customer cannot credit themselves by asking for more than they paid.
//
// Amount is in IDR. It is converted to quota units with
// config.PanWalletQuotaPerIDR.
func SyncWalletTopUp(c *gin.Context) {
	if walletClient() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Integrasi PanWallet belum diaktifkan",
		})
		return
	}

	link, err := model.GetWalletLinkByUserId(c.GetInt(ctxkey.Id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Wallet belum ditautkan"})
		return
	}

	var body struct {
		AmountIDR int64 `json:"amount_idr"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if body.AmountIDR <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "amount_idr harus positif"})
		return
	}

	bal, err := walletClient().GetBalance(link.PublicID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if bal.Available < body.AmountIDR {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("Saldo wallet tidak cukup. Tersedia Rp%d, diminta Rp%d",
				bal.Available, body.AmountIDR),
		})
		return
	}

	quota := int64(float64(body.AmountIDR) * config.PanWalletQuotaPerIDR)
	if quota <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Konversi ke kuota menghasilkan 0"})
		return
	}

	userId := c.GetInt(ctxkey.Id)
	if err := model.IncreaseUserQuota(userId, quota); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	model.RecordTopupLog(c.Request.Context(), userId,
		fmt.Sprintf("Top up via PanWallet Rp%d", body.AmountIDR), int(quota))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"quota":      quota,
			"amount_idr": body.AmountIDR,
			"balance":    bal.Available,
		},
	})
}

// AdminTopUpWallet credits a user's quota from a PanWallet wallet on the
// admin's behalf. Used when the customer paid but the self-service sync did
// not run (for example an offline QRIS invoice).
func AdminTopUpWallet(c *gin.Context) {
	if walletClient() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Integrasi PanWallet belum diaktifkan",
		})
		return
	}

	var req struct {
		UserId    int    `json:"user_id"`
		PublicID  string `json:"public_id"`
		AmountIDR int64  `json:"amount_idr"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if req.AmountIDR <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "amount_idr harus positif"})
		return
	}

	link, err := model.GetWalletLinkByUserId(req.UserId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "User belum punya wallet tertaut"})
		return
	}
	publicID := req.PublicID
	if publicID == "" {
		publicID = link.PublicID
	}

	bal, err := walletClient().GetBalance(publicID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if bal.Available < req.AmountIDR {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("Saldo wallet tidak cukup. Tersedia Rp%d", bal.Available),
		})
		return
	}

	quota := int64(float64(req.AmountIDR) * config.PanWalletQuotaPerIDR)
	if err := model.IncreaseUserQuota(req.UserId, quota); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	model.RecordTopupLog(c.Request.Context(), req.UserId,
		fmt.Sprintf("Top up via PanWallet (admin) Rp%d", req.AmountIDR), int(quota))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    gin.H{"quota": quota, "amount_idr": req.AmountIDR},
	})
}

// ListWalletLinks returns every wallet link, for the admin overview.
func ListWalletLinks(c *gin.Context) {
	links, err := model.ListWalletLinks()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": links})
}

var _ = errors.Is
var _ = strconv.Itoa
