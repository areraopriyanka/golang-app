package logging

import (
	"encoding/json"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ExampleStruct struct {
	Name     string
	Email    string `mask:"true"`
	Password string `mask:"true"`
}

func TestMaskValue(t *testing.T) {
	input := ExampleStruct{
		Name:     "John Doe",
		Email:    "user@example.com",
		Password: "Password123!",
	}

	expected := map[string]interface{}{
		"Name":     "John Doe",
		"Email":    "********",
		"Password": "********",
	}

	result := maskValue(input)

	assert.Equal(t, expected, result)
}

func TestMaskValueWithPointer(t *testing.T) {
	input := &ExampleStruct{
		Name:     "John Doe",
		Email:    "user@example.com",
		Password: "Password123!",
	}

	expected := map[string]interface{}{
		"Name":     "John Doe",
		"Email":    "********",
		"Password": "********",
	}

	result := maskValue(input)

	assert.Equal(t, expected, result)
}

// Redefining replaceAttr here since it is unexported in logger.go since it is
// scoped to the InitLogger function
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	env := os.Getenv("ENV")
	isProduction := env == "production"

	if !isProduction {
		return a
	}

	val := a.Value
	if val.Kind() == slog.KindAny {
		v := reflect.ValueOf(val.Any())

		if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
			return slog.Any(a.Key, maskValue(val.Any()))
		}
	}
	return a
}

func TestReplaceAttrInProduction(t *testing.T) {
	os.Setenv("ENV", "production")

	input := ExampleStruct{
		Name:     "John Doe",
		Email:    "user@example.com",
		Password: "Password123!",
	}

	attr := slog.Any("user", input)
	masked := replaceAttr(nil, attr)

	maskedVal := masked.Value.Any().(map[string]interface{})

	assert.Equal(t, "John Doe", maskedVal["Name"])
	assert.Equal(t, "********", maskedVal["Email"])
	assert.Equal(t, "********", maskedVal["Password"])
}

func TestObfuscateJsonFields(t *testing.T) {
	os.Setenv("ENV", "production")

	input := `{
		"id": "1",
		"result": {
			"api": {
				"type": "ACH_OUT_ACK",
				"reference": "ledger.ach.transfer_ach_out_1761772392647472835",
				"dateTime": "2025-10-29 21:13:13"
			},
			"account": {
				"accountId": "500400084011653",
				"balanceCents": 9665,
				"holdBalanceCents": 1269,
				"status": "ACTIVE"
			},
			"transactionNumber": "QA00000000048749",
			"transactionStatus": "COMPLETED",
			"transactionAmountCents": 1277,
			"originalRequestBase64": "eyJjaGFubmVsIjoiQUNIIiwidHJhbnNhY3Rpb25UeXBlIjoiQUNIX09VVCIsInRyYW5zYWN0aW9uRGF0ZVRpbWUiOiIyMDI1LTEwLTI5IDIxOjEzOjEyIiwicmVmZXJlbmNlIjoibGVkZ2VyLmFjaC50cmFuc2Zlcl9hY2hfb3V0XzE3NjE3NzIzOTI2NDc0NzI4MzUiLCJyZWFzb24iOiJBQ0ggQ3JlZGl0IiwidHJhbnNhY3Rpb25BbW91bnQiOnsiYW1vdW50IjoiMTI3NyIsImN1cnJlbmN5IjoiVVNEIn0sImRlYnRvciI6eyJmaXJzdE5hbWUiOiJjYXJzb24ifSwiZGVidG9yQWNjb3VudCI6eyJpZGVudGlmaWNhdGlvbiI6IjUwMDQwMDA4NDAxMTY1MyIsImlkZW50aWZpY2F0aW9uVHlwZSI6IkFDQ09VTlRfTlVNQkVSIiwiaW5zdGl0dXRpb24iOnsiaWRlbnRpZmljYXRpb24iOiIxMjQzMDMyOTgiLCJpZGVudGlmaWNhdGlvblR5cGUiOiJBQkEifX0sImNyZWRpdG9yIjp7InVzZXJUeXBlIjoiSU5ESVZJRFVBTCIsImZpcnN0TmFtZSI6IkFsYmVydGEgQm9iYmV0aCBDaGFybGVzb24ifSwiY3JlZGl0b3JBY2NvdW50Ijp7ImlkZW50aWZpY2F0aW9uIjoiMTExMTIyMjIzMzMzMDAwMCIsImlkZW50aWZpY2F0aW9uVHlwZSI6IkFDQ09VTlRfTlVNQkVSIiwiaWRlbnRpZmljYXRpb25UeXBlMiI6IkNIRUNLSU5HIiwiaW5zdGl0dXRpb24iOnsiaWRlbnRpZmljYXRpb24iOiIwMTE0MDE1MzMiLCJpZGVudGlmaWNhdGlvblR5cGUiOiJBQkEifX19",
			"processId": "PL25102900039329"
		},
		"header": {
			"reference": "ledger.ach.transfer_ach_out_1761772392647472835",
			"apiKey": "ce2b276bf19a4dea8d7b90f33f8a8892",
			"signature": "MEUCIGuG4Uzc05xSBYmJKN5jsM0Xsc7Y9FRlH2wYghrC7hMtAiEA0ZMh7j9XoP8Kwqa9U6U12z2QTDYtSx1awsM+OKOLvfA="
		}
	}`

	masked := maskJSONString(input)
	var result map[string]interface{}
	err := json.Unmarshal([]byte(masked), &result)
	assert.NoError(t, err)

	resultMap := result["result"].(map[string]interface{})
	assert.Equal(t, "********", resultMap["originalRequestBase64"])
	accountMap := resultMap["account"].(map[string]interface{})
	assert.Equal(t, "********", accountMap["accountId"])
	assert.Equal(t, "********", accountMap["balanceCents"])
	assert.Equal(t, "********", accountMap["holdBalanceCents"])

	headerMap := result["header"].(map[string]interface{})
	assert.Equal(t, "********", headerMap["apiKey"])
}
