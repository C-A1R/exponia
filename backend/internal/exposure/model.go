package exposure

import "time"

type Exposure struct {
	ID             int64      `json:"id"`
	FrameID        int64      `json:"frame_id"`
	ExposureIndex  int32      `json:"exposure_index"`
	CameraID       int64      `json:"camera_id"`
	LensID         *int64     `json:"lens_id"`
	Aperture       *float64   `json:"aperture"`
	ShutterSpeedUS *int64     `json:"shutter_speed_us"`
	ShotAt         *time.Time `json:"shot_at"`
	Note           *string    `json:"note"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateInput struct {
	LensID         *int64     `json:"lens_id"`
	Aperture       *float64   `json:"aperture"`
	ShutterSpeedUS *int64     `json:"shutter_speed_us"`
	ShotAt         *time.Time `json:"shot_at"`
	Note           *string    `json:"note"`
}

type UpdateInput struct {
	CameraID       int64      `json:"camera_id"`
	LensID         *int64     `json:"lens_id"`
	Aperture       *float64   `json:"aperture"`
	ShutterSpeedUS *int64     `json:"shutter_speed_us"`
	ShotAt         *time.Time `json:"shot_at"`
	Note           *string    `json:"note"`
}
