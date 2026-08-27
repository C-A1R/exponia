package camera

import "time"

type Camera struct {
	ID           int64     `json:"id"`
	Manufacturer string    `json:"manufacturer"`
	Model        string    `json:"model"`
	CreatedAt    time.Time `json:"created_at"`
}
