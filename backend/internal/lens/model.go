package lens

import "time"

type Lens struct {
	ID            int64     `json:"id"`
	Manufacturer  string    `json:"manufacturer"`
	Model         string    `json:"model"`
	FocalLengthMm int32     `json:"focal_length_mm"`
	MaxAperture   float64   `json:"max_aperture"`
	CreatedAt     time.Time `json:"created_at"`
}
