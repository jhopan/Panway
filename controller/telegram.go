package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/ctxkey"
	"github.com/jhopan/Panway/model"
)

// TelegramVerificationPurpose is the verification namespace for Telegram codes.
const TelegramVerificationPurpose = "t"

// telegramClient builds the Telegram client from config.
func telegramClient() *common.TelegramClient {
	if !config.TelegramVerificationEnabled {
		return nil
	}
	return common.NewTelegramClient(config.TelegramBotToken, config.TelegramApiBaseURL)
}

// SendTelegramVerificationCode asks the customer for their Telegram chat id
// and sends a verification code to it.
//
// The customer registers with email or Google first; Telegram only proves that
// they control a chat, so the code is bound to the logged-in user, not to an
// email address.
func SendTelegramVerificationCode(c *gin.Context) {
	client := telegramClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Verifikasi Telegram belum diaktifkan",
		})
		return
	}

	var body struct {
		ChatId string `json:"chat_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	chatId := common.ParseChatID(body.ChatId)
	if chatId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Chat ID Telegram tidak valid. Salin chat id dari @userinfobot",
		})
		return
	}

	// Make sure the chat exists and is reachable by the bot. Without this the
	// customer would wait for a code that can never arrive.
	if _, _, err := client.GetChat(chatId); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Chat Telegram tidak bisa dibaca: " + err.Error(),
		})
		return
	}

	userId := c.GetInt(ctxkey.Id)
	code := common.GenerateVerificationCode(6)
	common.RegisterVerificationCodeWithKey(telegramVerificationKey(userId, chatId), code, TelegramVerificationPurpose)

	botName := config.TelegramBotUsername
	if botName == "" {
		botName = "bot"
	}
	text := fmt.Sprintf(
		"Kode verifikasi %s kamu: <b>%s</b>\n\nKode berlaku %d menit.\nKalau kamu tidak meminta kode ini, abaikan saja.",
		config.SystemName, code, common.VerificationValidMinutes,
	)
	if err := client.SendText(chatId, text); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Gagal mengirim kode: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"chat_id":    chatId,
			"bot_name":   botName,
			"expires_in": common.VerificationValidMinutes * 60,
		},
	})
}

// VerifyTelegramCode checks the code the customer received in Telegram.
func VerifyTelegramCode(c *gin.Context) {
	if !config.TelegramVerificationEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Verifikasi Telegram belum diaktifkan",
		})
		return
	}

	var body struct {
		ChatId string `json:"chat_id"`
		Code   string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	chatId := common.ParseChatID(body.ChatId)
	if chatId == "" || strings.TrimSpace(body.Code) == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Chat ID dan kode wajib diisi",
		})
		return
	}

	userId := c.GetInt(ctxkey.Id)
	key := telegramVerificationKey(userId, chatId)
	if !common.VerifyCodeWithKey(key, strings.TrimSpace(body.Code), TelegramVerificationPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Kode salah atau sudah kedaluwarsa",
		})
		return
	}
	common.DeleteKey(key, TelegramVerificationPurpose)

	if err := model.SetTelegramVerified(userId, chatId); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    gin.H{"chat_id": chatId, "verified": true},
	})
}

// UnverifyTelegram removes the Telegram verification of the logged-in user.
func UnverifyTelegram(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	if err := model.ClearTelegramVerified(userId); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// GetTelegramStatus reports the verification state of the logged-in user.
func GetTelegramStatus(c *gin.Context) {
	user, err := model.GetUserById(c.GetInt(ctxkey.Id), false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	data := gin.H{
		"enabled":  config.TelegramVerificationEnabled,
		"verified": user.TelegramVerified,
		"chat_id":  user.TelegramChatId,
	}
	if config.TelegramVerificationEnabled && config.TelegramBotUsername != "" {
		data["bot_username"] = config.TelegramBotUsername
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": data})
}

// telegramVerificationKey binds a verification code to both the user and the
// chat, so a code sent to one chat cannot be used for another.
func telegramVerificationKey(userId int, chatId string) string {
	return fmt.Sprintf("%d:%s", userId, chatId)
}
