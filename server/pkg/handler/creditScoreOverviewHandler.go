package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"process-api/pkg/validators"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary CreditScoreOverviewHandler
// @Description Reports debtwise onboarding status and most recent credit score
// @Tags credit score
// @Accept json
// @Produce json
// @Success 200 {object} CreditScoreOverviewResponse
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/credit_score_overview [get]
func CreditScoreOverviewHandler(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(cc)

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	var score *int
	var increase *int
	isUserOnboarded := user.DebtwiseCustomerNumber != nil && user.DebtwiseOnboardingStatus == "complete"

	if isUserOnboarded {
		latestScore, err := dao.UserCreditScoreDao{}.FindLatestUserCreditScoreByUserId(userId)
		if err != nil {
			logger.Error("Error querying latest credit score", "error", err.Error())
			return response.InternalServerError(fmt.Sprintf("Error querying latest credit score: error: %s", err.Error()), errtrace.Wrap(err))
		}
		if latestScore != nil {
			jsonString, err := utils.DecryptKmsBinary(latestScore.EncryptedCreditData)
			if err != nil {
				logging.Logger.Error("Failed to decrypt encrypted credit data for credit score overview", "err", err)
			}

			var creditData dao.EncryptableCreditScoreData
			if err := json.Unmarshal([]byte(jsonString), &creditData); err != nil {
				logging.Logger.Error("Failed to unmarhsal encrypted credit data", "err", err)
			}

			if err := validators.ValidateStruct(creditData); err != nil {
				logging.Logger.Error("Failed to validate EncryptableCreditScoreData", "err", err)
			}

			score = &creditData.Score
			increase = &creditData.Increase
		}
	}

	return cc.JSON(http.StatusOK, CreditScoreOverviewResponse{
		IsUserOnboarded: isUserOnboarded,
		Score:           score,
		Increase:        increase,
	})
}

type CreditScoreOverviewResponse struct {
	IsUserOnboarded bool `json:"isUserOnboarded" validate:"required" mask:"true"`
	Score           *int `json:"score,omitempty" mask:"true"`
	Increase        *int `json:"increase,omitempty" mask:"true"`
}
