package maintenance

import (
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"time"
)

func DeleteExpiredOTPs() {
	// Calculate the expiration threshold
	expirationThreshold := clock.Now().Add(-time.Millisecond * time.Duration(config.Config.Otp.OtpExpiryDuration))

	// Perform the delete operation
	result := db.DB.Where("created_at < ?", expirationThreshold).Delete(&dao.UserOtpDao{})
	if result.Error != nil {
		logging.Logger.Error("Error while deleting expired OTPs from db", "error", result.Error)
		return
	}

	logging.Logger.Info("Count of deleted OTPs", "rowsAffected", result.RowsAffected)
}

func DeleteLedgerTokenRecords() {
	// Calculate the cutoff time for 60 minutes(3600000ms) ago
	cutoff := clock.Now().Add(-time.Duration(config.Config.Jwt.LedgerTokenExpTime) * time.Millisecond)
	logging.Logger.Info("Cutoff time in DeleteLedgerTokenRecords()", "cutoffTime", cutoff)
	// Delete ledger token records older than 60 minutes(3600000ms)
	result := db.DB.Where("created_at < ?", cutoff).Delete(&dao.LedgerUserDao{})
	if result.Error != nil {
		logging.Logger.Error("Error while deleting older notification from db", "error", result.Error)
		return
	}
	logging.Logger.Info("Count of deleted records", "rowsAffected", result.RowsAffected)
}
