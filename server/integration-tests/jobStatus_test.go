package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/handler"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestJobStatus() {
	h := suite.newHandler()

	sessionId := suite.SetupTestData()
	userPublicKey := suite.createUserPublicKeyRecord(sessionId)
	e := handler.NewEcho()

	ctx := context.Background()
	jobInsertResult, err := h.RiverClient.Insert(ctx, handler.CreditScoreJobArgs{
		UserId: sessionId,
	}, nil)

	suite.Require().NoError(err, "Failed to insert sort job")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/status/job/%d", jobInsertResult.Job.ID), nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/status/job/:jobId")
	c.SetParamNames("jobId")
	c.SetParamValues(fmt.Sprintf("%d", jobInsertResult.Job.ID))

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, userPublicKey.PublicKey, c)

	err = h.JobStatusHandler(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var statusResponseBody handler.JobStatusResponse

	err = json.Unmarshal(rec.Body.Bytes(), &statusResponseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	allowedStates := []string{
		"available", "cancelled", "completed", "discarded", "pending", "retryable", "running", "scheduled",
	}
	suite.Contains(allowedStates, statusResponseBody.State, "Job state should be of one of River's allowed states.")

	suite.WaitForJobsDone(1)
}
