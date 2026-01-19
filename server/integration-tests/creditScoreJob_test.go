package test

import (
	"encoding/json"
	"process-api/pkg/db/dao"
	"process-api/pkg/utils"
	"process-api/pkg/validators"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestCreditScoreJobQuery() {
	userRecord1 := suite.createTestUser(PartialMasterUserRecordDao{Id: utils.Pointer(uuid.New().String()), DebtwiseCustomerNumber: utils.Pointer(1), DebtwiseOnboardingStatus: utils.Pointer("complete"), Email: utils.Pointer("test1@example.com"), MobileNo: utils.Pointer("6140001234")})
	userRecord2 := suite.createTestUser(PartialMasterUserRecordDao{Id: utils.Pointer(uuid.New().String()), DebtwiseCustomerNumber: utils.Pointer(2), DebtwiseOnboardingStatus: utils.Pointer("complete"), Email: utils.Pointer("test2@example.com"), MobileNo: utils.Pointer("6150001234")})
	userRecord3 := suite.createTestUser(PartialMasterUserRecordDao{Id: utils.Pointer(uuid.New().String()), DebtwiseCustomerNumber: utils.Pointer(3), DebtwiseOnboardingStatus: utils.Pointer("complete"), Email: utils.Pointer("test3@example.com"), MobileNo: utils.Pointer("6160001234")})

	newCreditScoreId1 := uuid.New().String()

	encryptableCreditData := dao.EncryptableCreditScoreData{
		Score:                   550,
		Increase:                -5,
		DebtwiseCustomerNumber:  1,
		PaymentHistoryAmount:    97,
		PaymentHistoryFactor:    "Good",
		CreditUtilizationAmount: 10,
		CreditUtilizationFactor: "Good",
		DerogatoryMarksAmount:   2,
		DerogatoryMarksFactor:   "Good",
		CreditAgeAmount:         5,
		CreditAgeFactor:         "Good",
		CreditMixAmount:         3,
		CreditMixFactor:         "Good",
		NewCreditAmount:         1,
		NewCreditFactor:         "Good",
		TotalAccountsAmount:     6,
		TotalAccountsFactor:     "Good",
	}

	jsonData, err := json.Marshal(encryptableCreditData)
	suite.Require().NoError(err, "json marshalling should not return error")

	encryptedCreditData, err := utils.EncryptKmsBinary(string(jsonData))
	suite.Require().NoError(err, "encryption call should not return error")

	// User 1 New Score -- no update
	newDate, _ := time.Parse("2006-01-02", "2025-07-11")
	newCreditScoreRecord1 := dao.UserCreditScoreDao{
		Id:                  newCreditScoreId1,
		Date:                newDate,
		EncryptedCreditData: encryptedCreditData,
		UserId:              userRecord1.Id,
	}

	err = suite.TestDB.Create(&newCreditScoreRecord1).Error
	suite.Require().NoError(err, "Failed to insert new test credit report 1")

	oldCreditScoreId1 := uuid.New().String()

	// User 1 Old score -- superceded by new record
	oldDate, _ := time.Parse("2006-01-02", "2025-05-11")
	oldCreditScoreRecord1 := dao.UserCreditScoreDao{
		Id:                  oldCreditScoreId1,
		Date:                oldDate,
		EncryptedCreditData: encryptedCreditData,
		UserId:              userRecord1.Id,
	}

	err = suite.TestDB.Create(&oldCreditScoreRecord1).Error
	suite.Require().NoError(err, "Failed to insert old test credit report 1")

	oldCreditScoreId2 := uuid.New().String()

	// User 2 old score - should be updated
	oldCreditScoreRecord2 := dao.UserCreditScoreDao{
		Id:                  oldCreditScoreId2,
		Date:                oldDate,
		EncryptedCreditData: encryptedCreditData,
		UserId:              userRecord2.Id,
	}

	err = suite.TestDB.Create(&oldCreditScoreRecord2).Error
	suite.Require().NoError(err, "Failed to insert old test credit report 2")

	oldCreditScoreId3 := uuid.New().String()

	// User 2 old score 2 -- should be superceded by "newer" old score
	oldDate2, _ := time.Parse("2006-01-02", "2025-04-11")
	oldCreditScoreRecord3 := dao.UserCreditScoreDao{
		Id:                  oldCreditScoreId3,
		Date:                oldDate2,
		EncryptedCreditData: encryptedCreditData,
		UserId:              userRecord2.Id,
	}

	err = suite.TestDB.Create(&oldCreditScoreRecord3).Error
	suite.Require().NoError(err, "Failed to insert old test credit report 3")

	newCreditScoreId2 := uuid.New().String()

	// User 3 new score -- should not update
	newCreditScoreRecord2 := dao.UserCreditScoreDao{
		Id:                  newCreditScoreId2,
		Date:                newDate,
		EncryptedCreditData: encryptedCreditData,
		UserId:              userRecord3.Id,
	}

	err = suite.TestDB.Create(&newCreditScoreRecord2).Error
	suite.Require().NoError(err, "Failed to insert new test credit report 2")

	// Hardcoding this cutoff date to be 6-12-2025. The newest records have the Date value
	// of 7-11-2025 so this is simulating the behavior in the job for the following day
	// where it takes today's date and subtracts 30 days for the query cutoff date
	cutoffDate, err := time.Parse("2006-01-02", "2025-06-12")
	suite.Require().NoError(err, "Failed to parse hardcoded cutoff date")

	oldRecords, err := dao.UserCreditScoreDao{}.FindSufficientlyOldCreditScores(cutoffDate)
	suite.Require().NoError(err, "Query should not fail")

	jsonString, err := utils.DecryptKmsBinary(oldRecords[0].EncryptedCreditData)
	suite.Require().NoError(err, "Failed to decrypt encrypted credit data")

	var creditData dao.EncryptableCreditScoreData
	err = json.Unmarshal([]byte(jsonString), &creditData)
	suite.Require().NoError(err, "Failed to unmarhsal encrypted credit data")

	err = validators.ValidateStruct(creditData)
	suite.Require().NoError(err, "Failed to validate EncryptableCreditScoreData")
	suite.Require().Equal(creditData.CreditAgeAmount, 5.0, "Decrypted data must match")

	suite.Require().Equal(1, len(oldRecords), "Expected 1 record in total for users needing an update")

	suite.Require().Equal(userRecord2.Id, oldRecords[0].UserId, "User 2 is the only user without a sufficiently new record, so it is the only user record returned from the query")
}
