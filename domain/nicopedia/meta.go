package nicopedia

import "time"

// MetaData is nicopedia meta context info .
type MetaData struct {
	IsRedirect bool
	FromTitle  string
	CreateAt   time.Time
}
