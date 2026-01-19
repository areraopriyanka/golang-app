package ledger

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const oldRequestBody = `{
	"id": "35123",
	"createdDate": "2025-07-02T01:17:17.934Z",
	"updatedDate": "2025-07-02T01:17:17.934Z",
	"customerId": "100000000029028",
	"accountId": "21281104",
	"pdfV2File": "aGVsbG8K",
	"md5": "d41d8cd98f00b204e9800998ecf8427e",
	"accountNumber": "500400009957614",
	"accountName": "Default DreamFi Account",
	"customerName": "Test User",
	"legalRepID": [
	{
		"ID": "21281103",
		"name": "Test User"
	}
	],
	"currency": "USD",
	"month": "June",
	"year": 2025,
	"lastDate": "2025-07-01T03:59:59.999Z",
	"fileType": "PDFV2"
}`

func TestPdfBase64ForManyPayloadShapes(t *testing.T) {
	assertPayloadWithPdfVersion(t, oldRequestBody, "pdfFile")
	assertPayloadWithPdfVersion(t, oldRequestBody, "pdfV2File")
	assertPayloadWithPdfVersion(t, oldRequestBody, "pdfV3File")
	assertPayloadWithPdfVersion(t, oldRequestBody, "pdfV4File")
	assertPayloadWithPdfVersion(t, oldRequestBody, "pdfV5File")
	assertPayloadWithPdfVersion(t, oldRequestBody, "pdfV6File")
}

func assertPayloadWithPdfVersion(t *testing.T, oldRequestBody string, name string) {
	resultText := strings.Replace(oldRequestBody, "pdfV2File", name, 1)

	var result GetStatementResult
	err := json.Unmarshal([]byte(resultText), &result)
	assert.NoError(t, err, "Unmarshal shouldn't fail")

	base64, err := result.PdfBase64()
	assert.NoError(t, err)
	assert.Equal(t, "aGVsbG8K", base64, "PdfBase64 should resolve to '%s'", name)
}

func TestPdfBase64BadPayloadShape(t *testing.T) {
	resultText := strings.Replace(oldRequestBody, "pdfV2File", "aNameWeCanNotPredict", 1)

	var result GetStatementResult
	err := json.Unmarshal([]byte(resultText), &result)
	assert.NoError(t, err, "Unmarshal shouldn't fail")

	_, err = result.PdfBase64()
	assert.Error(t, err)
}
