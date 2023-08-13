package domain

import "time"

type User struct {
	Id           int64
	Email        string
	Password     string
	Nickname     string
	Birthday     string
	Introduction string
	Utime        time.Time
}
