package utils

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"regexp"
	"time"

	"braces.dev/errtrace"
	"github.com/dongri/phonenumber"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	lookup "github.com/twilio/twilio-go/rest/lookups/v2"
)

var CustomClient *customClient

type customClient struct {
	client     client.Client
	baseURL    string
	accountSID string
}

func InitializeTwilioClient(twilioConfig config.TwilioConfigs) {
	// Initializing custom twilio client
	CustomClient = &customClient{
		client: client.Client{
			Credentials: client.NewCredentials(twilioConfig.AccountSid, twilioConfig.AuthToken),
		},
		baseURL:    twilioConfig.ApiBase,
		accountSID: twilioConfig.AccountSid,
	}
}

func (c *customClient) SendRequest(method string, rawURL string, data url.Values, headers map[string]interface{}, body ...byte) (*http.Response, error) {
	logging.Logger.Info("Twilio default URL", "rawUrl", rawURL)

	// If ApiBase is set then override it else twilio will use default twilio ApiBase "https://api.twilio.com"
	if c.baseURL != "" {
		logging.Logger.Info("Overriding twilio APIBase with", "apiBase", config.Config.Twilio.ApiBase)
		baseMatcher := regexp.MustCompile(`^https?://[^/]+`)
		rawURL = baseMatcher.ReplaceAllString(rawURL, c.baseURL)
	}

	resp, err := c.client.SendRequest(method, rawURL, data, headers)

	return resp, errtrace.Wrap(err)
}

func (c *customClient) AccountSid() string {
	return c.accountSID
}

func (c *customClient) SetTimeout(timeout time.Duration) {
	if c.client.HTTPClient != nil {
		c.client.HTTPClient.Timeout = timeout
	} else {
		logging.Logger.Error("Twilio client initialization failed: HTTPClient is nil.")
	}
}

func SendSMS(to, body, from string) error {
	twilioAPIService := openapi.NewApiServiceWithClient(CustomClient)

	messageParams := &openapi.CreateMessageParams{}
	messageParams.SetTo(to)
	messageParams.SetFrom(from)
	messageParams.SetBody(body)

	_, err := twilioAPIService.CreateMessage(messageParams)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}

func MakeCall(to, from, twimlUrl string) error {
	twilioAPIService := openapi.NewApiServiceWithClient(CustomClient)
	callParams := &openapi.CreateCallParams{}
	callParams.SetTo(to)
	callParams.SetFrom(from)
	callParams.SetUrl(twimlUrl)
	callParams.SetStatusCallback(config.Config.Twilio.CallbackUrl)
	callParams.SetStatusCallbackEvent([]string{"initiated", "ringing", "answered", "completed", "failed", "busy", "no-answer"})
	_, err := twilioAPIService.CreateCall(callParams)
	if err != nil {
		return errtrace.Wrap(err)
	}

	return nil
}

func RemoveCountryCodeFromMobileNumber(phoneNumber string) (string, error) {
	parsedNumber := phonenumber.Parse(phoneNumber, "US")

	if parsedNumber == "" {
		return "", errtrace.Wrap(fmt.Errorf("only US phone numbers are valid. received: %s", phoneNumber))
	}

	if len(parsedNumber) > 10 {
		parsedNumber = parsedNumber[len(parsedNumber)-10:]
	}

	return parsedNumber, nil
}

func IsTwilioNonRequestError(err error) bool {
	var urlErr *url.Error

	// Cannot use direct type assertion here because the error is wrapped
	// errors.As() correctly unwraps and matches the *client.TwilioRestError

	if errors.As(err, &urlErr) {
		var netErr net.Error
		if errors.As(urlErr.Err, &netErr) {
			if netErr.Timeout() {
				logging.Logger.Error("Network timeout error while accessing Twilio service", "error", err.Error())
				return true
			}
		}
	}

	// Server side errors
	var twilioError *client.TwilioRestError
	if errors.As(err, &twilioError) {
		if twilioError.Status >= 500 {
			logging.Logger.Error(
				"Error while accessing Twilio server",
				"error", err.Error(),
				"statusCode", twilioError.Status,
			)
			return true
		}
	}

	return false
}

func IsValidMobileNumber(twilioConfig config.TwilioConfigs, phoneNumber string) (bool, error) {
	Client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: twilioConfig.AccountSid,
		Password: twilioConfig.AuthToken,
	})

	// Referred: https://www.twilio.com/docs/lookup/v2-api
	// Twilio's Lookup API returns a 'valid' boolean field indicating if the phone number is valid
	params := &lookup.FetchPhoneNumberParams{}
	phoneNumber = "+1" + phoneNumber
	resp, err := Client.LookupsV2.FetchPhoneNumber(phoneNumber, params)
	if err != nil {
		return false, errtrace.Wrap(err)
	}

	if resp != nil && resp.Valid != nil && *resp.Valid {
		return true, nil
	}

	return false, nil
}

func HandleTwilioError(err error) error {
	if IsTwilioNonRequestError(err) {
		return response.InternalServerError(fmt.Sprintf("Received non-request error from Twilio: %s", err.Error()), errtrace.Wrap(err))
	}
	// Handle: "Status: 400 - ApiError 21211: Invalid 'To' Phone Number: 401-112-XXXX (null) More info: https://www.twilio.com/docs/errors/21211"

	// Cannot use direct type assertion here because the error is wrapped
	// errors.As() correctly unwraps and matches the *client.TwilioRestError
	var twilioError *client.TwilioRestError
	if errors.As(err, &twilioError) {
		if twilioError.Status == 400 {
			// TODO: this should turn into response.BadRequestErrors
			return response.ErrorResponse{
				ErrorCode:       constant.INVALID_MOBILE_NUMBER,
				Message:         constant.INVALID_MOBILE_NUMBER_MSG,
				StatusCode:      http.StatusBadRequest,
				LogMessage:      fmt.Sprintf("Received 400 error from Twilio: %s", err.Error()),
				MaybeInnerError: errtrace.Wrap(err),
			}
		}
	}
	return response.InternalServerError(fmt.Sprintf("Received 4xx error from Twilio: %s", err.Error()), errtrace.Wrap(err))
}
