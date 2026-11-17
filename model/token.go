package model

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/helper"
	"github.com/jhopan/Panway/common/logger"
	"github.com/jhopan/Panway/common/message"
)

const (
	TokenStatusEnabled   = 1 // don't use 0, 0 is the default value!
	TokenStatusDisabled  = 2 // also don't use 0
	TokenStatusExpired   = 3
	TokenStatusExhausted = 4
)

type Token struct {
	Id             int     `json:"id"`
	UserId         int     `json:"user_id"`
	Key            string  `json:"key" gorm:"type:text"`
	KeyHash        string  `json:"-" gorm:"type:char(64);index"`
	Status         int     `json:"status" gorm:"default:1"`
	Name           string  `json:"name" gorm:"index" `
	CreatedTime    int64   `json:"created_time" gorm:"bigint"`
	AccessedTime   int64   `json:"accessed_time" gorm:"bigint"`
	ExpiredTime    int64   `json:"expired_time" gorm:"bigint;default:-1"` // -1 means never expired
	RemainQuota    int64   `json:"remain_quota" gorm:"bigint;default:0"`
	UnlimitedQuota bool    `json:"unlimited_quota" gorm:"default:false"`
	UsedQuota      int64   `json:"used_quota" gorm:"bigint;default:0"` // used quota
	Models         *string `json:"models" gorm:"type:text"`            // allowed models
	Subnet         *string `json:"subnet" gorm:"default:''"`           // allowed subnet
	// Rate limit per token. 0 = pakai default global.
	RateLimitRPM     int `json:"rate_limit_rpm" gorm:"default:0"`    // request per menit
	RateLimitTPM     int `json:"rate_limit_tpm" gorm:"default:0"`    // token per menit
	ConcurrencyLimit int `json:"concurrency_limit" gorm:"default:0"` // request paralel
}

// DecryptedKey returns the plaintext API key for the console. Values written
// before encryption was enabled are returned as-is.
func (t *Token) DecryptedKey() string {
	if t == nil {
		return ""
	}
	plain, err := common.DecryptTokenKey(t.Key)
	if err != nil {
		return ""
	}
	return plain
}

func GetAllUserTokens(userId int, startIdx int, num int, order string) ([]*Token, error) {
	var tokens []*Token
	var err error
	query := DB.Where("user_id = ?", userId)

	switch order {
	case "remain_quota":
		query = query.Order("unlimited_quota desc, remain_quota desc")
	case "used_quota":
		query = query.Order("used_quota desc")
	default:
		query = query.Order("id desc")
	}

	err = query.Limit(num).Offset(startIdx).Find(&tokens).Error
	return tokens, err
}

func SearchUserTokens(userId int, keyword string) (tokens []*Token, err error) {
	err = DB.Where("user_id = ?", userId).Where("name LIKE ?", keyword+"%").Find(&tokens).Error
	return tokens, err
}

func ValidateUserToken(key string) (token *Token, err error) {
	if key == "" {
		return nil, errors.New("No token provided")
	}
	hash, hashErr := common.TokenKeyHash(key)
	if hashErr == nil {
		token, err = CacheGetTokenByKeyHash(hash)
	}
	if hashErr != nil || err != nil {
		// A token created before key encryption has no hash yet. Look it up
		// by plaintext and migrate it in place.
		token, err = CacheGetTokenByKey(key)
		if err == nil {
			migrateTokenKey(token, key)
		}
	}
	if err != nil {
		logger.SysError("CacheGetTokenByKey failed: " + err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Invalid token")
		}
		return nil, errors.New("Token verification failed")
	}
	if token.Status == TokenStatusExhausted {
		return nil, fmt.Errorf("The quota of token %s (#%d) has been used up", token.Name, token.Id)
	} else if token.Status == TokenStatusExpired {
		return nil, errors.New("This token has expired")
	}
	if token.Status != TokenStatusEnabled {
		return nil, errors.New("This token's status is unavailable")
	}
	if token.ExpiredTime != -1 && token.ExpiredTime < helper.GetTimestamp() {
		if !common.RedisEnabled {
			token.Status = TokenStatusExpired
			err := token.SelectUpdate()
			if err != nil {
				logger.SysError("failed to update token status" + err.Error())
			}
		}
		return nil, errors.New("This token has expired")
	}
	if !token.UnlimitedQuota && token.RemainQuota <= 0 {
		if !common.RedisEnabled {
			// in this case, we can make sure the token is exhausted
			token.Status = TokenStatusExhausted
			err := token.SelectUpdate()
			if err != nil {
				logger.SysError("failed to update token status" + err.Error())
			}
		}
		return nil, errors.New("This token's quota has been used up")
	}
	return token, nil
}

func GetTokenByIds(id int, userId int) (*Token, error) {
	if id == 0 || userId == 0 {
		return nil, errors.New("id or userId is empty!")
	}
	token := Token{Id: id, UserId: userId}
	var err error = nil
	err = DB.First(&token, "id = ? and user_id = ?", id, userId).Error
	return &token, err
}

func GetTokenById(id int) (*Token, error) {
	if id == 0 {
		return nil, errors.New("id is empty!")
	}
	// Do not pre-set the primary key: GORM then appends its own primary-key
	// condition to the query, and the combination never matches a row.
	var token Token
	err := DB.Where("id = ?", id).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (t *Token) Insert() error {
	if t.Key != "" && !common.IsEncryptedTokenKey(t.Key) {
		hash, err := common.TokenKeyHash(t.Key)
		if err != nil {
			return err
		}
		encrypted, err := common.EncryptTokenKey(t.Key)
		if err != nil {
			return err
		}
		t.Key = encrypted
		t.KeyHash = hash
	}
	var err error
	err = DB.Create(t).Error
	return err
}

// migrateTokenKey encrypts one legacy plaintext key and records its lookup
// hash. It is called after a successful plaintext lookup and from MigrateTokenKeys.
func migrateTokenKey(token *Token, plain string) {
	if token == nil || plain == "" {
		return
	}
	hash, err := common.TokenKeyHash(plain)
	if err != nil {
		logger.SysError("token key hash failed: " + err.Error())
		return
	}
	encrypted, err := common.EncryptTokenKey(plain)
	if err != nil {
		logger.SysError("token key encryption failed: " + err.Error())
		return
	}
	err = DB.Model(&Token{}).Where("id = ?", token.Id).Updates(map[string]interface{}{
		"key":      encrypted,
		"key_hash": hash,
	}).Error
	if err != nil {
		logger.SysError("token key migration failed: " + err.Error())
		return
	}
	token.Key = encrypted
	token.KeyHash = hash
}

// MigrateTokenKeys encrypts every API key that was stored before key
// encryption was enabled. Startup calls it after the schema migration.
func MigrateTokenKeys() {
	var tokens []Token
	if err := DB.Where("key_hash = '' OR key_hash IS NULL").Find(&tokens).Error; err != nil {
		logger.SysError("token key migration query failed: " + err.Error())
		return
	}
	migrated := 0
	for i := range tokens {
		if tokens[i].Key == "" {
			continue
		}
		plain := tokens[i].Key
		if common.IsEncryptedTokenKey(plain) {
			decrypted, err := common.DecryptTokenKey(plain)
			if err != nil {
				logger.SysError("cannot decrypt token during migration: " + err.Error())
				continue
			}
			plain = decrypted
		}
		migrateTokenKey(&tokens[i], plain)
		migrated++
	}
	if migrated > 0 {
		logger.SysLogf("encrypted %d legacy token keys", migrated)
	}
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (t *Token) Update() error {
	var err error
	err = DB.Model(t).Select("name", "status", "expired_time", "remain_quota", "unlimited_quota", "models", "subnet",
		"rate_limit_rpm", "rate_limit_tpm", "concurrency_limit").Updates(t).Error
	return err
}

func (t *Token) SelectUpdate() error {
	// This can update zero values
	return DB.Model(t).Select("accessed_time", "status").Updates(t).Error
}

func (t *Token) Delete() error {
	var err error
	err = DB.Delete(t).Error
	return err
}

func (t *Token) GetModels() string {
	if t == nil {
		return ""
	}
	if t.Models == nil {
		return ""
	}
	return *t.Models
}

func DeleteTokenById(id int, userId int) (err error) {
	// Why we need userId here? In case user want to delete other's token.
	if id == 0 || userId == 0 {
		return errors.New("id or userId is empty!")
	}
	token := Token{Id: id, UserId: userId}
	err = DB.Where(token).First(&token).Error
	if err != nil {
		return err
	}
	return token.Delete()
}

func IncreaseTokenQuota(id int, quota int64) (err error) {
	if quota < 0 {
		return errors.New("quota cannot be negative!")
	}
	if config.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeTokenQuota, id, quota)
		return nil
	}
	return increaseTokenQuota(id, quota)
}

func increaseTokenQuota(id int, quota int64) (err error) {
	err = DB.Model(&Token{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"remain_quota":  gorm.Expr("remain_quota + ?", quota),
			"used_quota":    gorm.Expr("used_quota - ?", quota),
			"accessed_time": helper.GetTimestamp(),
		},
	).Error
	return err
}

func DecreaseTokenQuota(id int, quota int64) (err error) {
	if quota < 0 {
		return errors.New("quota cannot be negative!")
	}
	if config.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeTokenQuota, id, -quota)
		return nil
	}
	return decreaseTokenQuota(id, quota)
}

func decreaseTokenQuota(id int, quota int64) (err error) {
	err = DB.Model(&Token{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"remain_quota":  gorm.Expr("remain_quota - ?", quota),
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"accessed_time": helper.GetTimestamp(),
		},
	).Error
	return err
}

func PreConsumeTokenQuota(tokenId int, quota int64) (err error) {
	if quota < 0 {
		return errors.New("quota cannot be negative!")
	}
	if quota == 0 {
		return nil
	}
	token, err := GetTokenById(tokenId)
	if err != nil {
		return err
	}
	if !token.UnlimitedQuota && token.RemainQuota < quota {
		return errors.New("Insufficient token quota")
	}
	userQuota, err := GetUserQuota(token.UserId)
	if err != nil {
		return err
	}
	if userQuota < quota {
		return errors.New("Insufficient user quota")
	}
	quotaTooLow := userQuota >= config.QuotaRemindThreshold && userQuota-quota < config.QuotaRemindThreshold
	noMoreQuota := userQuota-quota <= 0
	if quotaTooLow || noMoreQuota {
		go func() {
			email, err := GetUserEmail(token.UserId)
			if err != nil {
				logger.SysError("failed to fetch user email: " + err.Error())
			}
			prompt := "Quota Reminder"
			var contentText string
			if noMoreQuota {
				contentText = "Your quota has been used up"
			} else {
				contentText = "Your quota is about to run out"
			}
			if email != "" {
				topUpLink := fmt.Sprintf("%s/topup", config.ServerAddress)
				content := message.EmailTemplate(
					prompt,
					fmt.Sprintf(`
						<p>Hello!</p>
						<p>%s, your remaining quota is <strong>%d</strong>.</p>
						<p>To avoid any interruption in your usage, please top up in time.</p>
						<p style="text-align: center; margin: 30px 0;">
							<a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Top Up Now</a>
						</p>
						<p style="color: #666;">If the button cannot be clicked, copy the following link and open it in your browser.</p>
						<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px; word-break: break-all;">%s</p>
					`, contentText, userQuota, topUpLink, topUpLink),
				)
				err = message.SendEmail(prompt, email, content)
				if err != nil {
					logger.SysError("failed to send email: " + err.Error())
				}
			}
		}()
	}
	if !token.UnlimitedQuota {
		// Reserve the token's own budget in one conditional UPDATE. Reading
		// the balance and writing it back in two steps lets two concurrent
		// requests both pass the check and spend the same quota.
		reserved, err := reserveTokenQuota(tokenId, quota)
		if err != nil {
			return err
		}
		if !reserved {
			return errors.New("Insufficient token quota")
		}
	}
	err = DecreaseUserQuota(token.UserId, quota)
	if err != nil {
		// Hand the token reservation back so a failed user charge does not
		// silently eat the customer's token balance.
		if !token.UnlimitedQuota {
			if refundErr := IncreaseTokenQuota(tokenId, quota); refundErr != nil {
				logger.SysError("failed to refund token quota: " + refundErr.Error())
			}
		}
		return err
	}
	return nil
}

// reserveTokenQuota takes quota from a limited token only when the balance
// still covers it. It reports false when the balance is too low.
func reserveTokenQuota(id int, quota int64) (bool, error) {
	result := DB.Model(&Token{}).
		Where("id = ? AND unlimited_quota = ?", id, false).
		Where("remain_quota >= ?", quota).
		Updates(map[string]interface{}{
			"remain_quota":  gorm.Expr("remain_quota - ?", quota),
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"accessed_time": helper.GetTimestamp(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// consumeRemainingTokenQuota charges at most the token's remaining balance.
//
// The pre-consumed estimate can be smaller than the real usage (the upstream
// may add its own prompt, or the response may be longer than max_tokens), so
// the post-consume step must not push the balance below zero. The CASE
// expressions read the old row value in every supported database.
func consumeRemainingTokenQuota(id int, quota int64) error {
	return DB.Exec(`
		UPDATE tokens
		SET used_quota = used_quota + CASE WHEN remain_quota >= ? THEN ? ELSE remain_quota END,
		    remain_quota = CASE WHEN remain_quota >= ? THEN remain_quota - ? ELSE 0 END,
		    accessed_time = ?
		WHERE id = ? AND unlimited_quota = ?`,
		quota, quota, quota, quota, helper.GetTimestamp(), id, false,
	).Error
}

func PostConsumeTokenQuota(tokenId int, quota int64) (err error) {
	token, err := GetTokenById(tokenId)
	if err != nil {
		return err
	}
	if quota > 0 {
		err = DecreaseUserQuota(token.UserId, quota)
	} else if quota < 0 {
		err = IncreaseUserQuota(token.UserId, -quota)
	}
	if err != nil {
		return err
	}
	if !token.UnlimitedQuota {
		if quota > 0 {
			// Charge at most the remaining balance. The pre-consumed
			// estimate can be smaller than the real usage, and a negative
			// token balance would let a request slip through on the next
			// replay of the quota check.
			err = consumeRemainingTokenQuota(tokenId, quota)
		} else if quota < 0 {
			err = IncreaseTokenQuota(tokenId, -quota)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
