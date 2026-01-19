package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/logging"

	"braces.dev/errtrace"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendEmail(recipientName, recipientEmail, subject, messageData string) error {
	sendGridClient := sendgrid.NewSendClient(config.Config.Email.ApiKey)

	if config.Config.Email.ApiBase != "" {
		logging.Logger.Info("Overriding sendgrid's BaseUrl with", "BaseUrl", config.Config.Email.ApiBase)
		// NewSendClient constructor adds /v3/mail/send endpoint to the default BaseURL (https://api.sendgrid.com) in case of Sending email.
		// So the final BaseUrl https://api.sendgrid.com/v3/mail/send
		// If we override the BaseURL for a mock server, we need to manually add /v3/mail/send.
		sendGridClient.BaseURL = config.Config.Email.ApiBase + "/v3/mail/send"
	}
	from := mail.NewEmail("DreamFi", config.Config.Email.FromAddr)
	to := mail.NewEmail(recipientName, recipientEmail)
	message := mail.NewSingleEmail(from, subject, to, "", messageData)

	response, err := sendGridClient.Send(message)
	if err != nil {
		logging.Logger.Error("Error while sending email", "error", err)
		return errtrace.Wrap(err)
	}
	// Sendgrid returns 202(http.StatusAccepted) on success
	if response.StatusCode != http.StatusAccepted {
		logging.Logger.Error("Error while sending email", "statusCode", response.StatusCode, "Body", response.Body)
		return errtrace.Wrap(fmt.Errorf("error: Sendgrid returned non-success response"))
	}

	return nil
}

func GenerateEmailBody(templateFileName string, data interface{}) (string, error) {
	// Read the HTML template from the specified file
	tmpl, err := template.ParseFiles(templateFileName)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to load template: %s", err.Error()))
	}
	// Buffer to store the executed template output
	var buffer bytes.Buffer
	// Adding dynamic data in template
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to execute template: %s", err.Error()))
	}
	return buffer.String(), nil
}
