package model

// SetTelegramVerified marks the user's Telegram chat as verified.
func SetTelegramVerified(userId int, chatId string) error {
	return DB.Model(&User{}).Where("id = ?", userId).
		Updates(map[string]any{
			"telegram_chat_id":  chatId,
			"telegram_verified": true,
		}).Error
}

// ClearTelegramVerified removes the Telegram verification of a user.
func ClearTelegramVerified(userId int) error {
	return DB.Model(&User{}).Where("id = ?", userId).
		Updates(map[string]any{
			"telegram_chat_id":  "",
			"telegram_verified": false,
		}).Error
}

// GetTelegramVerified reports the Telegram verification state of a user.
func GetTelegramVerified(userId int) (chatId string, verified bool, err error) {
	user, err := GetUserById(userId, false)
	if err != nil {
		return "", false, err
	}
	return user.TelegramChatId, user.TelegramVerified, nil
}
