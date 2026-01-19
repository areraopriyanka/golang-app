package response

type OtpEmailTemplateData struct {
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	OtpExpTimeMinutes string `json:"otpExpTime"`
	ImageData         string `json:"imageData"`
	Otp               string `json:"otp"`
}

type StatementEmailTemplateData struct {
	FirstName string `json:"firstName"`
	Month     string `json:"month"`
	Year      string `json:"year"`
	AppLink   string `json:"appLink"`
}
