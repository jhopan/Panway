package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/jhopan/Panway/common/logger"
)

const (
	PlanStatusEnabled  = 1
	PlanStatusDisabled = 2
)

// Plan is a prepaid package a customer can buy with their balance.
//
// PriceIDR and Quota use the same unit as quota (Rp1 by default). Quota is
// credited when the plan is bought and PriceIDR is debited; the controller
// rejects a plan that would credit more than it costs, otherwise buying it
// repeatedly would create balance out of nothing. DurationDays controls when
// the plan access expires, and Models optionally restricts the models the
// customer may call while the plan is active.
type Plan struct {
	Id           int    `json:"id"`
	Name         string `json:"name" gorm:"unique;index"`
	Description  string `json:"description" gorm:"type:text"`
	PriceIDR     int64  `json:"price_idr" gorm:"column:price_idr;bigint;default:0"`
	Quota        int64  `json:"quota" gorm:"bigint;default:0"`
	DurationDays int    `json:"duration_days" gorm:"default:30"`
	Models       string `json:"models" gorm:"type:text"`
	Status       int    `json:"status" gorm:"default:1"`
	CreatedTime  int64  `json:"created_time" gorm:"bigint"`
}

func ListActivePlans() ([]*Plan, error) {
	var plans []*Plan
	err := DB.Where("status = ?", PlanStatusEnabled).Order("price_idr asc").Find(&plans).Error
	return plans, err
}

func ListAllPlans() ([]*Plan, error) {
	var plans []*Plan
	err := DB.Order("id desc").Find(&plans).Error
	return plans, err
}

func GetPlanById(id int) (*Plan, error) {
	if id <= 0 {
		return nil, errors.New("plan id is empty")
	}
	var plan Plan
	if err := DB.First(&plan, id).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (p *Plan) Insert() error {
	p.CreatedTime = time.Now().Unix()
	return DB.Create(p).Error
}

func (p *Plan) UpdateFields() error {
	return DB.Model(&Plan{}).Where("id = ?", p.Id).Updates(map[string]interface{}{
		"name":          p.Name,
		"description":   p.Description,
		"price_idr":     p.PriceIDR,
		"quota":         p.Quota,
		"duration_days": p.DurationDays,
		"models":        p.Models,
		"status":        p.Status,
	}).Error
}

func DeletePlan(id int) error {
	return DB.Delete(&Plan{}, id).Error
}

// GetActiveUserPlan returns the plan attached to a user, or nil when the user
// has no plan or it has expired.
func GetActiveUserPlan(userId int) (*Plan, error) {
	var user User
	if err := DB.Select("plan_id", "plan_expires_at").First(&user, userId).Error; err != nil {
		return nil, err
	}
	if user.PlanId == 0 || user.PlanExpiresAt <= time.Now().Unix() {
		return nil, nil
	}
	return GetPlanById(user.PlanId)
}

// BuyPlan debits the plan price and credits its quota and expiry in one
// transaction. It returns the user's quota after the purchase.
func BuyPlan(userId int, plan *Plan) (int64, int64, error) {
	var newQuota int64
	var expiresAt int64
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Select("id", "quota", "plan_expires_at").First(&user, userId).Error; err != nil {
			return err
		}
		if plan.PriceIDR > 0 && user.Quota < plan.PriceIDR {
			return errors.New("insufficient quota to buy this plan")
		}
		if plan.DurationDays <= 0 {
			return errors.New("this plan has no duration")
		}
		start := time.Now().Unix()
		if user.PlanExpiresAt > start {
			start = user.PlanExpiresAt
		}
		expiresAt = start + int64(plan.DurationDays)*24*60*60

		res := tx.Model(&User{}).
			Where("id = ? AND quota >= ?", userId, plan.PriceIDR).
			Updates(map[string]interface{}{
				"quota":           gorm.Expr("quota - ? + ?", plan.PriceIDR, plan.Quota),
				"plan_id":         plan.Id,
				"plan_expires_at": expiresAt,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("insufficient quota to buy this plan")
		}
		if err := tx.Model(&User{}).Where("id = ?", userId).Pluck("quota", &newQuota).Error; err != nil {
			return err
		}
		return tx.Create(&Log{
			UserId:    userId,
			CreatedAt: time.Now().Unix(),
			Type:      LogTypeTopup,
			Content:   "Plan: " + plan.Name,
			Quota:     int(plan.Quota),
		}).Error
	})
	if err != nil {
		return 0, 0, err
	}
	return newQuota, expiresAt, nil
}

// StartPlanExpiry clears expired plans once a day. The plan check already
// ignores expired values, but clearing them keeps the user row honest and
// makes the console show the right state.
func StartPlanExpiry(ctx context.Context) {
	go func() {
		expirePlans()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				expirePlans()
			}
		}
	}()
}

func expirePlans() int64 {
	now := time.Now().Unix()
	res := DB.Model(&User{}).
		Where("plan_id <> 0 AND plan_expires_at > 0 AND plan_expires_at <= ?", now).
		Updates(map[string]interface{}{"plan_id": 0})
	if res.Error != nil {
		logger.SysError("plan expiry failed: " + res.Error.Error())
		return 0
	}
	if res.RowsAffected > 0 {
		logger.SysLog(fmt.Sprintf("expired plans for %d users", res.RowsAffected))
	}
	return res.RowsAffected
}
