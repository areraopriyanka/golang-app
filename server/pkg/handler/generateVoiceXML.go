package handler

import (
	"net/http"
	"process-api/pkg/config"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func GenerateVoiceXML(c echo.Context) error {
	otp := c.QueryParam("otp")
	expirationInMinutes := strconv.Itoa(config.Config.Otp.OtpExpiryDuration / 60000)
	formattedOTP := strings.Join(strings.Split(otp, ""), " ")
	response := `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Say voice="Polly.Joanna">
        <prosody rate="slow">
			<break time="500ms"/>
            Your One Time Password to verify your phone number is
			<break time="500ms"/>
			<emphasis level="strong">` + formattedOTP + `</emphasis>
			<break time="500ms"/>
			This One Time Password will be valid for the next ` + expirationInMinutes + ` minute(s).
			<break time="400ms"/>
			Do not share this with anyone
			<break time="400ms"/>
			Thank you.
        </prosody>
    </Say>
</Response>`

	return c.Blob(http.StatusOK, "application/xml", []byte(response))
}
