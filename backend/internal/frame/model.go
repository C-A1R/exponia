package frame

import "time"

type Frame struct {
	ID         int64     `json:"id"`
	FilmRollID int64     `json:"film_roll_id"`
	FrameIndex int32     `json:"frame_index"`
	FrameLabel *string   `json:"frame_label"`
	Note       *string   `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}
