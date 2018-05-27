package nicopedia

import "time"

type MetaData struct {
	IsRedirect bool
	FromTitle  string
	CreateAt   time.Time
}
