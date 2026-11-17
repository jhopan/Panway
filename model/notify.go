package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/logger"
)

// lowBalanceLastSent throttles reminders so a customer is not messaged on
// every pass of the notifier.
var lowBalanceLastSent = struct {
	sync.Mutex
	byUser map[int]time.Time
}{byUser: map[int]time.Time{}}

// StartLowBalanceNotifier sends a Telegram reminder when a verified customer's
// quota falls below LOW_BALANCE_THRESHOLD.
//
// Telegram is already the verification channel, so it is the cheapest way to
// reach a customer without adding another provider. Users without a verified
// chat are skipped; email would need SMTP configured and is not sent here.
func StartLowBalanceNotifier(ctx context.Context) {
	if config.LowBalanceThreshold <= 0 {
		return
	}
	client := common.NewTelegramClient(config.TelegramBotToken, config.TelegramApiBaseURL)
	if !client.Enabled() {
		logger.SysError("LOW_BALANCE_THRESHOLD is set but Telegram is not configured; skipping low-balance reminders")
		return
	}
	interval := time.Duration(config.LowBalanceNotifyIntervalHours) * time.Hour
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		notifyLowBalance(client, interval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				notifyLowBalance(client, interval)
			}
		}
	}()
	logger.SysLogf("low-balance Telegram reminders enabled below %d quota", config.LowBalanceThreshold)
}

func notifyLowBalance(client *common.TelegramClient, interval time.Duration) {
	var users []User
	err := DB.Where("status = ? AND telegram_verified = ? AND telegram_chat_id <> '' AND quota < ?",
		UserStatusEnabled, true, config.LowBalanceThreshold).Find(&users).Error
	if err != nil {
		logger.SysError("low-balance query failed: " + err.Error())
		return
	}
	now := time.Now()
	for _, user := range users {
		lowBalanceLastSent.Lock()
		last, seen := lowBalanceLastSent.byUser[user.Id]
		if seen && now.Sub(last) < interval {
			lowBalanceLastSent.Unlock()
			continue
		}
		lowBalanceLastSent.byUser[user.Id] = now
		lowBalanceLastSent.Unlock()

		topUpURL := config.ServerAddress
		if topUpURL == "" {
			topUpURL = "http://localhost:3000"
		}
		text := fmt.Sprintf("Your Panway balance is low: %d quota left. Top up at %s/topup", user.Quota, topUpURL)
		if err := client.SendText(user.TelegramChatId, text); err != nil {
			logger.SysError(fmt.Sprintf("low-balance reminder to user %d failed: %s", user.Id, err.Error()))
		}
	}
}
