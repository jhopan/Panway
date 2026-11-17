package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/ctxkey"
	"github.com/jhopan/Panway/model"
)

// ListPlans returns the active plan catalogue to a signed-in customer.
func ListPlans(c *gin.Context) {
	plans, err := model.ListActivePlans()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": plans})
}

// ListAllPlans returns every plan (including disabled ones) to an admin.
func ListAllPlans(c *gin.Context) {
	plans, err := model.ListAllPlans()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": plans})
}

func GetPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	plan, err := model.GetPlanById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": plan})
}

func CreatePlan(c *gin.Context) {
	var plan model.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request body"})
		return
	}
	if err := normalizeAndValidatePlan(&plan); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := plan.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	auditf(c, "create plan #%d %s", plan.Id, plan.Name)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": plan})
}

func UpdatePlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	existing, err := model.GetPlanById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	var plan model.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request body"})
		return
	}
	plan.Id = existing.Id
	if err := normalizeAndValidatePlan(&plan); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := plan.UpdateFields(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	auditf(c, "update plan #%d %s", plan.Id, plan.Name)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": plan})
}

func DeletePlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := model.DeletePlan(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	auditf(c, "delete plan #%d", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// BuyPlan charges the plan price from the customer's balance and extends the
// plan expiry. The plan validation guarantees that a plan never credits more
// quota than it costs, so repeated purchases cannot create balance.
func BuyPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	plan, err := model.GetPlanById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if plan.Status != model.PlanStatusEnabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "this plan is not available"})
		return
	}
	userId := c.GetInt(ctxkey.Id)
	quota, expiresAt, err := model.BuyPlan(userId, plan)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.CacheUpdateUserQuota(c.Request.Context(), userId); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"quota":           quota,
			"plan_id":         plan.Id,
			"plan_name":       plan.Name,
			"plan_expires_at": expiresAt,
		},
	})
}

func normalizeAndValidatePlan(plan *model.Plan) error {
	plan.Name = strings.TrimSpace(plan.Name)
	plan.Description = strings.TrimSpace(plan.Description)
	plan.Models = strings.TrimSpace(plan.Models)
	if plan.Name == "" {
		return errors.New("plan name is required")
	}
	if len(plan.Name) > 60 {
		return errors.New("plan name is too long")
	}
	if plan.PriceIDR < 0 || plan.Quota < 0 {
		return errors.New("price and quota cannot be negative")
	}
	if plan.PriceIDR > 0 && plan.Quota > plan.PriceIDR {
		return errors.New("a plan cannot credit more quota than it costs")
	}
	if plan.DurationDays <= 0 || plan.DurationDays > 3650 {
		return errors.New("duration must be between 1 and 3650 days")
	}
	if plan.Status != model.PlanStatusEnabled && plan.Status != model.PlanStatusDisabled {
		plan.Status = model.PlanStatusEnabled
	}
	if plan.Models != "" {
		for _, item := range strings.Split(plan.Models, ",") {
			if strings.TrimSpace(item) == "" {
				return errors.New("model list contains an empty entry")
			}
		}
	}
	return nil
}
