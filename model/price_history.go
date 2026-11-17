package model

import (
	"time"
)

// PriceHistory records an admin price change so the catalogue can show a
// short trend instead of only the current number.
type PriceHistory struct {
	Id          int     `json:"id"`
	ModelId     string  `json:"model_id" gorm:"type:varchar(191);index"`
	InputPrice  float64 `json:"input_price"`
	OutputPrice float64 `json:"output_price"`
	CreatedAt   int64   `json:"created_at" gorm:"bigint;index"`
}

func RecordPriceHistory(modelId string, inputPrice, outputPrice float64) error {
	if modelId == "" {
		return nil
	}
	return DB.Create(&PriceHistory{
		ModelId:     modelId,
		InputPrice:  inputPrice,
		OutputPrice: outputPrice,
		CreatedAt:   time.Now().Unix(),
	}).Error
}

func GetPriceHistorySince(since int64) ([]PriceHistory, error) {
	var rows []PriceHistory
	err := DB.Where("created_at >= ?", since).Order("created_at asc").Find(&rows).Error
	return rows, err
}

// PrunePriceHistory keeps the table small; the catalogue only needs a few
// days of edits.
func PrunePriceHistory(keepDays int) {
	if keepDays <= 0 {
		return
	}
	cutoff := time.Now().Add(-time.Duration(keepDays) * 24 * time.Hour).Unix()
	DB.Where("created_at < ?", cutoff).Delete(&PriceHistory{})
}
