package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lmittmann/tint"
)

var (
	Logger     *slog.Logger
	projectDir string
)

func InitLogger() {
	// Create a new logger instance
	var handler slog.Handler

	_, path, _, _ := runtime.Caller(1)
	dir, err := filepath.Abs(filepath.Dir(path))
	projectDir = dir
	if err != nil {
		panic("Failed to find project working directory: " + err.Error())
	}

	env := os.Getenv("ENV")
	isProduction := env == "production"

	level := slog.LevelDebug
	if isProduction {
		level = slog.LevelInfo
	}

	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
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

		if val.Kind() == slog.KindString && json.Valid([]byte(val.String())) {
			masked := maskJSONString(val.String())
			return slog.String(a.Key, masked)
		}

		return a
	}

	logJson := os.Getenv("LOG_JSON")
	isJsonLog := logJson == "true"

	if isJsonLog {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   true,
			Level:       level,
			ReplaceAttr: replaceAttr,
		})
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:   true,
			TimeFormat:  time.RFC3339,
			Level:       level,
			ReplaceAttr: replaceAttr,
		},
		)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

func RequestContextLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		contextLogger := Logger.With(slog.String("req", fmt.Sprintf("%s %s", c.Request().Method, c.Request().URL.Path)))
		c.Set("contextLogger", contextLogger)

		if contextLogger.Handler().Enabled(context.Background(), slog.LevelDebug) {
			reqBody := []byte{}
			if c.Request().Body != nil { // Read
				reqBody, _ = io.ReadAll(c.Request().Body)
			}
			c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))

			if len(reqBody) > 0 {
				var readableBody bytes.Buffer
				if err := json.Indent(&readableBody, reqBody, "", "  "); err != nil {
					readableBody = *bytes.NewBuffer(reqBody)
				}
				contextLogger.Debug("Request body", slog.String("request", readableBody.String()))
			} else {
				contextLogger.Debug("Request")
			}
		} else {
			contextLogger.Info("Request")
		}

		return next(c)
	}
}

func RequestResponseBodyLogger(c echo.Context, reqBody, resBody []byte) {
	contextLogger := GetEchoContextLogger(c)

	if len(resBody) > 0 {
		var prettyRes bytes.Buffer
		if err := json.Indent(&prettyRes, resBody, "", "  "); err != nil {
			prettyRes = *bytes.NewBuffer(resBody)
		}
		contextLogger.Debug("Response body", slog.String("response", prettyRes.String()))
	} else {
		contextLogger.Debug("Response", "status", c.Response().Status)
	}
}

func GetEchoContextLogger(c echo.Context) *slog.Logger {
	if c == nil {
		Logger.Error("Called GetEchoContextLogger with nil echo Context")
		return Logger
	}

	contextLogger := c.Get("contextLogger")
	if logger, ok := contextLogger.(*slog.Logger); ok {
		return logger
	}

	Logger.Error("echo context is not set up with contextLogger or is of wrong type")
	return Logger
}

func maskValue(maskableStruct interface{}) interface{} {
	structValue := reflect.ValueOf(maskableStruct)
	structType := reflect.TypeOf(maskableStruct)

	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
		structType = structType.Elem()
	}

	if structValue.Kind() == reflect.Struct {
		maskedMap := make(map[string]interface{})
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			fieldValue := structValue.Field(i)

			// protect against unexported values like time
			if !fieldValue.CanInterface() {
				continue
			}

			if field.Tag.Get("mask") == "true" {
				maskedMap[field.Name] = "********"
			} else {
				maskedMap[field.Name] = fieldValue.Interface()
			}
		}
		return maskedMap
	}

	return maskableStruct
}

var piiFields = []string{
	"originalRequestBase64", "balanceCents", "holdBalanceCents", "accountId", "apiKey",
}

func maskJSONString(jsonStr string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr
	}

	maskJSONValue(data)

	masked, err := json.Marshal(data)
	if err != nil {
		return jsonStr
	}

	return string(masked)
}

func maskJSONValue(value interface{}) {
	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if isPIIField(key) {
				v[key] = "********"
			} else {
				maskJSONValue(val)
			}
		}
	case []interface{}:
		for _, item := range v {
			maskJSONValue(item)
		}
	}
}

func isPIIField(field string) bool {
	fieldLower := strings.ToLower(field)
	for _, pii := range piiFields {
		if fieldLower == strings.ToLower(pii) {
			return true
		}
	}
	return false
}

func removeProjectPath(path string) string {
	return strings.Replace(path, projectDir, "", 1)
}

type GormLogger struct {
	Logger *slog.Logger
}

func (gl *GormLogger) Print(v ...interface{}) {
	if gl == nil {
		return
	}

	switch len(v) {
	case 1: // error
		gl.Logger.Error("gorm error", slog.String("message", fmt.Sprintf("%v", v[0])))

	case 3: // error with file/line
		message := fmt.Sprintf("%v", v[2])
		filePath := removeProjectPath(fmt.Sprintf("%v", v[1]))

		gl.Logger.Error("gorm error at %s: %s", filePath, message)

	case 6:
		// NOTE: This breaks on gorm upgrade, but there are other libraries
		// s.print("sql", fileWithLineNum(), NowFunc().Sub(t), sql, vars, s.RowsAffected)
		//
		// Calling log here breaks the frame tracking so we manually append the file/line gorm calls with
		sql := fmt.Sprintf("%v", v[3])
		vars := fmt.Sprintf("%v", v[4])
		rowsAffected := fmt.Sprintf("%v", v[5])

		gl.Logger.With(
			slog.String("time", fmt.Sprintf("%v", v[2])),
			slog.String("vars", vars),
			slog.String("rows", rowsAffected),
		).Debug("gorm SQL executed", slog.String("sql", sql))

	default:
		gl.Logger.Error("Unexpected args length", slog.Int("args_length", len(v)), slog.Any("args", v))
	}
}
