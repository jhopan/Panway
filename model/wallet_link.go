package model

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// WalletLink ties a Panway user to a PanWallet wallet.
//
// PanWallet owns the money; Panway only mirrors the balance as quota units.
// The link is created by the customer: they enter their short PanWallet id
// (public_id) in the Panway panel, and Panway verifies it against the PanWallet
// API before storing it.
//
// PublicID is the customer's short id (8 characters, case-insensitive).
// WalletUserID is PanWallet's internal UUID. Both are stored so Panway can
// talk to either the public or the product-scoped endpoint.
type WalletLink struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id" gorm:"uniqueIndex"`
	PublicID     string `json:"public_id" gorm:"type:char(32);index"`
	WalletUserID string `json:"wallet_user_id" gorm:"type:char(64);index"`
	CreatedTime  int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime  int64  `json:"updated_time" gorm:"bigint"`
}

func GetWalletLinkByUserId(userId int) (*WalletLink, error) {
	if userId == 0 {
		return nil, errors.New("invalid user id")
	}
	link := WalletLink{}
	err := DB.First(&link, "user_id = ?", userId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWalletLinkNotFound
		}
		return nil, err
	}
	return &link, nil
}

func GetWalletLinkByPublicId(publicID string) (*WalletLink, error) {
	link := WalletLink{}
	err := DB.First(&link, "public_id = ?", publicID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWalletLinkNotFound
		}
		return nil, err
	}
	return &link, nil
}

// UpsertWalletLink creates or updates the link for a user.
func UpsertWalletLink(userId int, publicID, walletUserID string) error {
	now := time.Now().Unix()
	link := WalletLink{
		UserId:       userId,
		PublicID:     publicID,
		WalletUserID: walletUserID,
		CreatedTime:  now,
		UpdatedTime:  now,
	}
	return DB.Save(&link).Error
}

func DeleteWalletLink(userId int) error {
	return DB.Where("user_id = ?", userId).Delete(&WalletLink{}).Error
}

// ListWalletLinks returns every link, for the admin overview.
func ListWalletLinks() ([]*WalletLink, error) {
	var links []*WalletLink
	err := DB.Order("id desc").Find(&links).Error
	return links, err
}

// ErrWalletLinkNotFound is returned when no wallet is linked to the user.
var ErrWalletLinkNotFound = errors.New("wallet not linked")

var _ = context.Background
