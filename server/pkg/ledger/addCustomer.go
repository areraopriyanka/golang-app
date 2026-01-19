package ledger

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"process-api/pkg/validators"
	"strings"
	"time"

	"braces.dev/errtrace"
)

func AddCustomer(userId string, requestPayload AddCustomerData) (customerNumber string, err error) {
	logger := logging.Logger.With(
		slog.String("ledgerMethod", "AddCustomer"),
		slog.String("userId", userId),
	)
	payload := createPayload(requestPayload)

	// validate addCustomerPayload
	if err := validators.ValidateStruct(payload); err != nil {
		return "", errtrace.Wrap(err)
	}

	payloadDataJSON, marshallErr := json.Marshal(payload)
	if marshallErr != nil {
		logger.Error("Could not marshal payload", "error", marshallErr.Error())
		return "", errtrace.Wrap(marshallErr)
	}
	request := &Request{
		Id:     constant.LEDGER_REQUEST_ID,
		Method: "CustomerService.AddCustomer",
		Params: Params{
			Payload: payloadDataJSON,
		},
	}

	signingErr := request.SignRequestWithMiddlewareKey()
	if signingErr != nil {
		logger.Error("An error occured while signing request payload", "error", signingErr.Error())
		return "", errtrace.Wrap(signingErr)
	}

	statusCode, respBody, ledgerErr := CallLedgerAPIAndGetRawResponse(nil, request)
	if ledgerErr != nil {
		logger.Error("Ledger returned an error response", "error", ledgerErr.Error())
		return "", errtrace.Wrap(ledgerErr)
	}

	if statusCode != http.StatusOK {
		logger.Error("Ledger returned an error", "statusCode", statusCode, "respBody", string(respBody))
		return "", errtrace.Wrap(errors.New("ledger returned an error statusCode"))
	}

	var response addCustomerResponse
	unmarshalErr := json.Unmarshal(respBody, &response)
	if unmarshalErr != nil {
		logger.Error("An error occurred while unmarshaling ledger response", "error", unmarshalErr.Error())
		return "", errtrace.Wrap(unmarshalErr)
	}

	// The response error object can contain a `Code` field; e.g., `CUSTOMER_IDENTIFICATION_ALREADY_EXIST`.
	// The codes are not documented, but might be worth investigating if we see repeated errors here.
	if response.Error != nil {
		logger.Error("The ledger responded with an error", "code", response.Error.Code, "msg", response.Error.Message)
		return "", errtrace.Wrap(errors.New(response.Error.Code))
	}
	// The customer's status might not be active (could be e.g., IN_ACTIVE, DORMANT, etc.),
	// in which case we may want to return an error.
	return response.Result.CustomerNumber, nil
}

func createPayload(request AddCustomerData) addCustomerPayload {
	return addCustomerPayload{
		Type:      "INDIVIDUAL",
		DOB:       request.DOB.Format("20060102"), // (format: YYYYMMDD)
		Title:     "Ms",                           // NOTE: hardcoding Title, as we don't collect it
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Identification: []identification{{
			Type:  "SSN",
			Value: request.SSN,
		}},
		Contact: contact{
			PhoneNumber: strings.ReplaceAll(request.PhoneNumber, "-", ""),
			Email:       request.Email,
		},
		Address: Address{
			AddressLine1: request.AddressLine1,
			AddressLine2: request.AddressLine2,
			City:         request.City,
			State:        request.State,
			Country:      "US", // NOTE: hardcoding Country, as we don't collect it
			ZIP:          request.ZIP,
		},
		UserName: request.Email,
		Password: request.Password,
	}
}

type AddCustomerData struct {
	DOB          time.Time `json:"DOB"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	PhoneNumber  string    `json:"phoneNumber"`
	Email        string    `json:"email"`
	AddressLine1 string    `json:"addressLine1"`
	AddressLine2 string    `json:"addressLine2"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	ZIP          string    `json:"zip"`
	UserName     string    `json:"UserName"`
	Password     string    `json:"password"`
	SSN          string    `json:"ssn"`
}

type identification struct {
	Type  string `json:"type" validate:"required,oneof=SSN"`
	Value string `json:"value" validate:"required,validateSSN"`
}

type contact struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,validatePhone"`
	Email       string `json:"email" validate:"required,validateEmail"`
}

type Address struct {
	AddressLine1 string `json:"addressLine1" validate:"required,validateAddress"`
	// NOTE: The ledger does not yet describe this field. However, it is validated when a value is provided.
	AddressLine2 string `json:"addressLine2" validate:"validateAddress"`
	City         string `json:"city" validate:"required,validateCity"`
	State        string `json:"state" validate:"required,validateState"`
	Country      string `json:"country" validate:"required,oneof=US"`
	ZIP          string `json:"zip" validate:"required,validateZipCode"`
}

// Source: https://apidocs.netxd.com/developers/docs/customer_apis/Add_Customer_-_Consumer
type addCustomerPayload struct {
	Type string `json:"type" validate:"required"`
	DOB  string `json:"DOB" validate:"required,validateLedgerDOB"`
	// TODO: we don't collect this, but it's required by netxd; oneof=Mr Ms Mrs
	Title     string `json:"title" validate:"required"`
	FirstName string `json:"firstName" validate:"required,validateName"`
	LastName  string `json:"lastName" validate:"required,validateName"`
	// NOTE: Gender is one of the few optional fields; we don't collect that data
	Identification []identification `json:"identification" validate:"required,min=1"`
	Contact        contact          `json:"contact" validate:"required"`
	Address        Address          `json:"address" validate:"required"`
	// NOTE: UserName and Password are marked optional by netxd, but we need them for our purposes
	UserName string `json:"UserName" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Source: https://apidocs.netxd.com/developers/docs/customer_apis/Add_Customer_-_Consumer
type addCustomerResponseResult struct {
	Id             string `json:"Id"`
	CustomerNumber string `json:"CustomerNumber"`
	Status         string `json:"Status"`
}

// Source: directly calling the endpoint
type addCustomerResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type addCustomerResponse struct {
	Id     string                     `json:"id"` // validate:"required"
	Result *addCustomerResponseResult `json:"Result,omitempty"`
	Error  *addCustomerResponseError  `json:"error,omitempty"`
}
