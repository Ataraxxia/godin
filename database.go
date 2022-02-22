package main

import (
	"time"

	rep "github.com/Ataraxxia/godin/report"
)

type Database interface {
	InitializeDatabase() error
	SaveReport(r rep.Report, t time.Time) error
}
