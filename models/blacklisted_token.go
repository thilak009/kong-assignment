package models

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"github.com/thilak009/kong-assignment/db"
	"github.com/thilak009/kong-assignment/pkg/log"
	"github.com/thilak009/kong-assignment/utils"
	"gorm.io/gorm"
)

type BlacklistedToken struct {
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	TokenHash string    `gorm:"uniqueIndex;primaryKey"`
	UserID    string    `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
}

func (bt *BlacklistedToken) BeforeCreate(tx *gorm.DB) (err error) {
	bt.CreatedAt = time.Now()
	bt.UpdatedAt = time.Now()
	return
}

func (bt *BlacklistedToken) BeforeUpdate(tx *gorm.DB) (err error) {
	bt.UpdatedAt = time.Now()
	return
}

type BlacklistedTokenModel struct{}

// HashToken creates a SHA256 hash of the token for storage
func (m BlacklistedTokenModel) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// Create adds a token to the blacklist
func (m BlacklistedTokenModel) Create(ctx context.Context, tokenHash, userID string, expiresAt time.Time) error {
	db := db.GetDB()

	blacklistedToken := BlacklistedToken{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	if err := db.Create(&blacklistedToken).Error; err != nil {
		log.With(ctx).Errorf("failed to blacklist token for user with id %s :: error: %s", userID, err.Error())
		return err
	}

	return nil
}

// IsBlacklisted checks if a token hash is in the blacklist and not expired
func (m BlacklistedTokenModel) IsBlacklisted(ctx context.Context, tokenHash string) bool {
	db := db.GetDB()
	var count int64

	err := db.Model(&BlacklistedToken{}).
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		Count(&count).Error

	if err != nil {
		log.With(ctx).Errorf("failed to check if token is blacklisted :: error: %s", err.Error())
		// On error, assume token is valid to avoid blocking users
		return false
	}

	return count > 0
}

// CleanupExpired removes expired tokens from the blacklist
func (m BlacklistedTokenModel) CleanupExpired(ctx context.Context) error {
	db := db.GetDB()

	result := db.Where("expires_at <= ?", time.Now()).Delete(&BlacklistedToken{})
	if result.Error != nil {
		log.With(ctx).Errorf("failed to cleanup expired blacklisted tokens :: error: %s", result.Error.Error())
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.With(ctx).Infof("cleaned up %d expired blacklisted tokens", result.RowsAffected)
	}

	return nil
}

// StartTokenCleanup runs periodic cleanup of expired blacklisted tokens
func StartTokenCleanup() {
	logger := log.GetLogger()
	blacklistModel := BlacklistedTokenModel{}

	// Get cleanup interval from environment (default: 1 hour)
	cleanupIntervalHours, err := strconv.Atoi(utils.GetEnv("TOKEN_CLEANUP_INTERVAL_MINUTES", "60"))
	if err != nil || cleanupIntervalHours < 1 {
		cleanupIntervalHours = 1
	}

	cleanupInterval := time.Duration(cleanupIntervalHours) * time.Minute
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	logger.Infof("Started periodic token cleanup (runs every %d minute(s))", cleanupIntervalHours)

	for {
		select {
		case <-ticker.C:
			logger.Info("running token clean up")
			ctx := context.Background()
			if err := blacklistModel.CleanupExpired(ctx); err != nil {
				logger.Errorf("Failed to cleanup expired tokens: %s", err.Error())
			}
		}
	}
}
