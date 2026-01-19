package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/ledger"
	"process-api/pkg/security"

	"github.com/google/uuid"
)

const HardcodedCardId = "5434a20e0a9b4d1bbf8fea3f543f9a05"

func (suite *IntegrationTestSuite) TestActivateCard() {
	defer SetupMockForLedger(suite).Close()
	user := suite.createUser()
	encryptedCvv := "fbbVg8NQFDwa5rtkjR+Wa4wqiZwUWoxaUVLO2fTFgeYvQ9RGBfyl6H0rpCvtEoEbhbVFTuJGX8hdgjfptOynnrKD5DUFpK0RDi+URuKIunzbiQ9Gq6iqs54S2LaWH0ZVwsvPb4HiP1mOX88oV7xZc7rzWcG1nOBErogItPTijlG/UMvVWU0SYMd+fjBNqMcdMeFe3GLch3KMFpAC4yC+sTSyLj9W8C0vOhF2bA3voxX2Vb5Xv27oR8Xn6jKoR+/ZzK0Ch3aF06P+Xxto/ZGZyvdxsN+s2ldEu+vHbK6QwXC30/U+J8TFOQ0rIbZny3LOtaBK0sbflAdLNlQwaFmCfQ=="
	encryptedPin := "fbbVg8NQFDwa5rtkjR+Wa4wqiZwUWoxaUVLO2fTFgeYvQ9RGBfyl6H0rpCvtEoEbhbVFTuJGX8hdgjfptOynnrKD5DUFpK0RDi+URuKIunzbiQ9Gq6iqs54S2LaWH0ZVwsvPb4HiP1mOX88oV7xZc7rzWcG1nOBErogItPTijlG/UMvVWU0SYMd+fjBNqMcdMeFe3GLch3KMFpAC4yC+sTSyLj9W8C0vOhF2bA3voxX2Vb5Xv27oR8Xn6jKoR+/ZzK0Ch3aF06P+Xxto/ZGZyvdxsN+s2ldEu+vHbK6QwXC30/U+J8TFOQ0rIbZny3LOtaBK0sbflAdLNlQwaFmCfQ=="

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        user.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}
	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	userClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, nil)

	validateCvv, err := userClient.BuildValidateCvvRequest(HardcodedCardId, encryptedCvv, false)
	suite.Require().NoError(err, "BuildChangePinRequest should not return an error")
	validateCvvPinPayloadId := suite.createSignablePayload(validateCvv, user.Id)

	changePin, err := userClient.BuildChangePinRequest(user.LedgerCustomerNumber, userAccountCard.AccountNumber, HardcodedCardId, encryptedPin, false)
	suite.Require().NoError(err, "BuildChangePinRequest should not return an error")
	changePinPayloadId := suite.createSignablePayload(changePin, user.Id)

	updateStatusPayloadId := suite.createUpdateStatusSignablePayload(user.LedgerCustomerNumber, userAccountCard.AccountNumber, encryptedCvv, ledger.ACTIVATE_CARD, user.Id)

	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	activateCardRequest := handler.ActivateCardRequest{
		ValidateCvv: handler.PayloadRequest{
			Signature: "test_signature",
			PayloadId: validateCvvPinPayloadId,
		},
		SetCardPin: handler.PayloadRequest{
			Signature: "test_signature",
			PayloadId: changePinPayloadId,
		},
		UpdateStatus: handler.PayloadRequest{
			Signature: "test_signature",
			PayloadId: updateStatusPayloadId,
		},
	}
	requestBody, _ := json.Marshal(activateCardRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/activate", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/activate")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err = handler.ActivateCard(customContext)
	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")
}

func (suite *IntegrationTestSuite) createUpdateStatusSignablePayload(customerNumber, accountNumber, cvv, action, userId string) string {
	config.Config.Ledger.CardsPublicKey = "-----BEGIN PUBLIC KEY-----\nMIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgEgiYAXZyoZzdUhqkCrwNJxBtbPBoEVaOGSCko+IkDCR93UJzuzBIv3286IVM7xXUEpmIj9MKnebY5CgKb9hAv6kt1clhuNpPYWYRHU/uq/PH31fYL6yf/e7bG4YoAHu1Ov212oqjgejerbTZVyeel3AKPdVP9mGu4sqmXLa+QQXAgMBAAE=\n-----END PUBLIC KEY-----"
	userClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, nil)

	payloadData, err := userClient.BuildUpdateStatusRequest(customerNumber, "5434a20e0a9b4d1bbf8fea3f543f9a05", accountNumber, action, cvv, false)
	suite.Require().NoError(err, "Failed to generate update status payload")

	return suite.createSignablePayload(payloadData, userId)
}

func (suite *IntegrationTestSuite) createSignablePayload(payloadData any, userId string) string {
	jsonPayloadBytes, err := json.Marshal(payloadData)
	suite.Require().NoError(err, "Failed to marshal test payload")

	jsonPayload := string(jsonPayloadBytes)

	payloadId := uuid.New().String()
	payloadRecord := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: jsonPayload,
		UserId:  &userId,
	}
	err = suite.TestDB.Create(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to insert test payload")
	return payloadId
}
