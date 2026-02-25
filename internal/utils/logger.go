package utils

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "[GO-PARSER] ", log.LstdFlags|log.Lshortfile)
