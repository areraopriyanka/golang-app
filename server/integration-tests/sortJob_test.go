package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestStartSortJob() {
	h := suite.newHandler()

	sessionId := suite.SetupTestData()
	userPublicKey := suite.createUserPublicKeyRecord(sessionId)
	e := handler.NewEcho()

	sortRequest := handler.SortJobStartRequest{
		StringsToSort: []string{"Salem", "Yachats", "Klamath", "Tillamook"},
	}

	requestBody, err := json.Marshal(sortRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPut, "/start-job/start", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/sort-job/start")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, userPublicKey.PublicKey, c)

	err = h.StartSortJob(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.MaybeRiverJobResponse[handler.SortJobOutput]
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	jobId := responseBody.JobId

	suite.IsType(int(0), jobId, "Job ID is present and of type int")
	suite.Require().NotEmpty(responseBody.Payload.SortedStrings, "With polling behavior expect sort job to have completed sorting")
	suite.Require().Equal(responseBody.State, "completed", "With polling behavior expect sort job to have state completed")

	expectedSortedStrings := []string{"Klamath", "Salem", "Tillamook", "Yachats"}

	sortStatusRequest := request.JobStatusRequest{
		JobId: jobId,
	}

	statusRequestBody, err := json.Marshal(sortStatusRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	var statusResponseBody response.MaybeRiverJobResponse[handler.SortJobOutput]

	// NOTE: Polling the status endpoint until we receive job status completed
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/sort-job/status", bytes.NewReader(statusRequestBody))
	req2.Header.Set("Content-Type", "application/json")
	c2 := e.NewContext(req2, rec2)
	c2.SetPath("/sort-job/status")
	customContext2 := security.GenerateLoggedInRegisteredUserContext(sessionId, userPublicKey.PublicKey, c2)

	err = h.GetSortJobStatus(customContext2)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec2.Code, "Expected status code 200 OK")

	err = json.Unmarshal(rec2.Body.Bytes(), &statusResponseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.IsType("completed", statusResponseBody.State, "Response indicates the status of the job as completed")
	suite.Require().Equal(expectedSortedStrings, statusResponseBody.Payload.SortedStrings, "Output returns the properly sorted strings")
}
