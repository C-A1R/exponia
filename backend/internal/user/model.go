package user

import "time"

type User struct {
	ID          int64
	Email       string
	DisplayName string
	AuthIssuer  string
	AuthSubject string
	CreatedAt   time.Time
}
