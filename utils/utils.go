package utils

import (
	"fmt"
	"os"

	"github.com/jhunt/go-ansi"
)

func Bail(f string, args ...interface{}) {
	ansi.Fprintf(os.Stderr, "@R{"+f+"}\n", args...)
	os.Exit(1)
}

func CallerName() string {
	return fmt.Sprintf(
		"%s/%s/%s:%s",
		os.Getenv("BUILD_TEAM_NAME"),
		os.Getenv("BUILD_PIPELINE_NAME"),
		os.Getenv("BUILD_JOB_NAME"),
		os.Getenv("BUILD_NAME"),
	)
}
