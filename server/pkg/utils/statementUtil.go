package utils

import (
	"fmt"
	"time"

	"braces.dev/errtrace"
)

func GenerateStatementTitle(lastDate string) (string, error) {
	t, err := time.Parse(time.RFC3339, lastDate)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to parse lastDate: %s", lastDate))
	}
	statementTitle := "Statement " + t.Format("02/01/06")
	return statementTitle, nil
}

func GenerateStatementFileName(lastDate string) (string, error) {
	t, err := time.Parse(time.RFC3339, lastDate)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to parse lastDate: %s", lastDate))
	}
	fileName := fmt.Sprintf("DreamFi_Statement_%s.pdf", t.Format("2006-01-02"))
	return fileName, nil
}
