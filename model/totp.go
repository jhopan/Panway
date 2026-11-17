package model

import (
	"errors"

	"github.com/jhopan/Panway/common"
)

func SetUserTotpSecret(userId int, secret string) error {
	if userId == 0 {
		return errors.New("user id is empty")
	}
	encrypted, err := common.EncryptTokenKey(secret)
	if err != nil {
		return err
	}
	return DB.Model(&User{}).Where("id = ?", userId).Update("totp_secret", encrypted).Error
}

func GetUserTotpSecret(userId int) (string, error) {
	if userId == 0 {
		return "", errors.New("user id is empty")
	}
	var stored string
	err := DB.Model(&User{}).Where("id = ?", userId).Pluck("totp_secret", &stored).Error
	if err != nil {
		return "", err
	}
	return common.DecryptTokenKey(stored)
}

func GetUserTotpStatus(userId int) (bool, error) {
	if userId == 0 {
		return false, errors.New("user id is empty")
	}
	var enabled bool
	err := DB.Model(&User{}).Where("id = ?", userId).Pluck("totp_enabled", &enabled).Error
	return enabled, err
}

func SetUserTotpEnabled(userId int, enabled bool) error {
	if userId == 0 {
		return errors.New("user id is empty")
	}
	updates := map[string]interface{}{"totp_enabled": enabled}
	if !enabled {
		updates["totp_secret"] = ""
	}
	return DB.Model(&User{}).Where("id = ?", userId).Updates(updates).Error
}
