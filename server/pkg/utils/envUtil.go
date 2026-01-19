package utils

import "os"

func GetEnv() string {
	var env string
	if len(os.Args) > 1 {
		env = os.Args[1]
	} else {
		env = os.Getenv("ENV")
	}
	return env
}
