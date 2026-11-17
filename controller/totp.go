package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/ctxkey"
	"github.com/jhopan/Panway/model"
)

const pending2FAKey = "pending_2fa_user"

func GetTotpStatus(c *gin.Context) {
	enabled, err := model.GetUserTotpStatus(c.GetInt(ctxkey.Id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{"enabled": enabled}})
}

// SetupTotp generates a secret but does not enable 2FA yet. The client shows
// the secret/otpauth URL and calls EnableTotp with one code from the app.
func SetupTotp(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	enabled, err := model.GetUserTotpStatus(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "2FA is already enabled"})
		return
	}
	secret, err := common.GenerateTOTPSecret()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.SetUserTotpSecret(userId, secret); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"secret": secret,
			"url":    common.TOTPURL(c.GetString(ctxkey.Username), secret),
		},
	})
}

func EnableTotp(c *gin.Context) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request body"})
		return
	}
	userId := c.GetInt(ctxkey.Id)
	secret, err := model.GetUserTotpSecret(userId)
	if err != nil || secret == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "run the 2FA setup first"})
		return
	}
	if !common.ValidateTOTP(secret, body.Code, time.Now()) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid 2FA code"})
		return
	}
	if err := model.SetUserTotpEnabled(userId, true); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "2FA enabled"})
}

// DisableTotp requires the account password, so a stolen session alone cannot
// turn the second factor off.
func DisableTotp(c *gin.Context) {
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request body"})
		return
	}
	user, err := model.GetUserById(c.GetInt(ctxkey.Id), true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if !common.ValidatePasswordAndHash(body.Password, user.Password) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "incorrect password"})
		return
	}
	if err := model.SetUserTotpEnabled(user.Id, false); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "2FA disabled"})
}

// LoginTOTP completes the second step of password login.
func LoginTOTP(c *gin.Context) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request body"})
		return
	}
	session := sessions.Default(c)
	pendingId, _ := session.Get(pending2FAKey).(int)
	if pendingId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "no login is waiting for a 2FA code"})
		return
	}
	user, err := model.GetUserById(pendingId, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	secret, err := model.GetUserTotpSecret(user.Id)
	if err != nil || secret == "" || !common.ValidateTOTP(secret, body.Code, time.Now()) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid 2FA code"})
		return
	}
	session.Delete(pending2FAKey)
	_ = session.Save()
	SetupLogin(user, c)
}
