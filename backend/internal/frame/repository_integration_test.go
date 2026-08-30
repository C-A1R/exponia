package frame_test

import (
	"errors"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/filmroll"
	"github.com/C-A1R/exponia/backend/internal/frame"
	"github.com/C-A1R/exponia/backend/internal/testutil"
	"github.com/C-A1R/exponia/backend/internal/user"
)

func TestRepository(t *testing.T) {
	pool := testutil.StartPostgres(t)

	userRepository := user.NewRepository(pool)
	owner, err := userRepository.FindOrCreate(
		t.Context(),
		"frame-owner@example.com",
		"Frame Owner",
		"https://auth.example.com",
		"frame-owner",
	)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}

	otherUser, err := userRepository.FindOrCreate(
		t.Context(),
		"frame-other@example.com",
		"Frame Other User",
		"https://auth.example.com",
		"frame-other-user",
	)
	if err != nil {
		t.Fatalf("create other user: %v", err)
	}

	var filmStockID int64
	var formatID int64

	err = pool.QueryRow(
		t.Context(),
		`
		SELECT fs.id, ff.id
		FROM film_stocks fs
		JOIN film_stock_formats fsf
		    ON fsf.film_stock_id = fs.id
		JOIN film_formats ff
		    ON ff.id = fsf.format_id
		WHERE fs.manufacturer = 'Kodak'
		  AND fs.name = 'Gold 200'
		  AND ff.code = '35mm'
		`,
	).Scan(&filmStockID, &formatID)
	if err != nil {
		t.Fatalf("find test film stock and format: %v", err)
	}

	filmRollRepository := filmroll.NewRepository(pool)
	roll, err := filmRollRepository.Create(
		t.Context(),
		owner.ID,
		filmStockID,
		formatID,
	)
	if err != nil {
		t.Fatalf("create film roll: %v", err)
	}

	repository := frame.NewRepository(pool)
	label := "00"
	note := "First frame"

	first, err := repository.Create(
		t.Context(),
		owner.ID,
		roll.ID,
		&label,
		&note,
	)
	if err != nil {
		t.Fatalf("create first frame: %v", err)
	}

	if first.FrameIndex != 1 {
		t.Fatalf("expected first frame index 1, got %d", first.FrameIndex)
	}

	if first.FrameLabel == nil || *first.FrameLabel != label {
		t.Fatalf("expected frame label %q, got %v", label, first.FrameLabel)
	}

	if first.Note == nil || *first.Note != note {
		t.Fatalf("expected note %q, got %v", note, first.Note)
	}

	second, err := repository.Create(
		t.Context(),
		owner.ID,
		roll.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("create second frame: %v", err)
	}

	if second.FrameIndex != 2 {
		t.Fatalf("expected second frame index 2, got %d", second.FrameIndex)
	}

	if second.FrameLabel != nil {
		t.Fatalf("expected nil frame label, got %q", *second.FrameLabel)
	}

	if second.Note != nil {
		t.Fatalf("expected nil note, got %q", *second.Note)
	}

	fetched, err := repository.GetByID(
		t.Context(),
		owner.ID,
		first.ID,
	)
	if err != nil {
		t.Fatalf("get owner's frame: %v", err)
	}

	if fetched.ID != first.ID {
		t.Fatalf("expected frame id %d, got %d", first.ID, fetched.ID)
	}

	_, err = repository.GetByID(
		t.Context(),
		otherUser.ID,
		first.ID,
	)
	if !errors.Is(err, frame.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	updatedLabel := "00A"
	updatedNote := "Updated first frame"
	updated, err := repository.Update(
		t.Context(),
		owner.ID,
		first.ID,
		&updatedLabel,
		&updatedNote,
	)
	if err != nil {
		t.Fatalf("update owner's frame: %v", err)
	}

	if updated.FrameIndex != first.FrameIndex {
		t.Fatalf(
			"expected frame index to remain %d, got %d",
			first.FrameIndex,
			updated.FrameIndex,
		)
	}

	if updated.FrameLabel == nil || *updated.FrameLabel != updatedLabel {
		t.Fatalf(
			"expected updated label %q, got %v",
			updatedLabel,
			updated.FrameLabel,
		)
	}

	if updated.Note == nil || *updated.Note != updatedNote {
		t.Fatalf(
			"expected updated note %q, got %v",
			updatedNote,
			updated.Note,
		)
	}

	cleared, err := repository.Update(
		t.Context(),
		owner.ID,
		first.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("clear frame fields: %v", err)
	}

	if cleared.FrameLabel != nil || cleared.Note != nil {
		t.Fatalf(
			"expected cleared fields, got label %v and note %v",
			cleared.FrameLabel,
			cleared.Note,
		)
	}

	_, err = repository.Update(
		t.Context(),
		otherUser.ID,
		second.ID,
		&updatedLabel,
		&updatedNote,
	)
	if !errors.Is(err, frame.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	unchanged, err := repository.GetByID(
		t.Context(),
		owner.ID,
		second.ID,
	)
	if err != nil {
		t.Fatalf("get frame after forbidden update: %v", err)
	}

	if unchanged.FrameLabel != nil || unchanged.Note != nil {
		t.Fatalf(
			"expected frame to remain unchanged, got label %v and note %v",
			unchanged.FrameLabel,
			unchanged.Note,
		)
	}

	frames, err := repository.List(
		t.Context(),
		owner.ID,
		roll.ID,
	)
	if err != nil {
		t.Fatalf("list owner's frames: %v", err)
	}

	if len(frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames))
	}

	if frames[0].FrameIndex != 1 || frames[1].FrameIndex != 2 {
		t.Fatalf(
			"expected frame indexes [1, 2], got [%d, %d]",
			frames[0].FrameIndex,
			frames[1].FrameIndex,
		)
	}

	_, err = repository.List(
		t.Context(),
		otherUser.ID,
		roll.ID,
	)
	if !errors.Is(err, frame.ErrFilmRollNotFound) {
		t.Fatalf("expected ErrFilmRollNotFound, got %v", err)
	}

	_, err = repository.List(
		t.Context(),
		owner.ID,
		999999,
	)
	if !errors.Is(err, frame.ErrFilmRollNotFound) {
		t.Fatalf("expected ErrFilmRollNotFound, got %v", err)
	}

	emptyRoll, err := filmRollRepository.Create(
		t.Context(),
		owner.ID,
		filmStockID,
		formatID,
	)
	if err != nil {
		t.Fatalf("create empty film roll: %v", err)
	}

	emptyFrames, err := repository.List(
		t.Context(),
		owner.ID,
		emptyRoll.ID,
	)
	if err != nil {
		t.Fatalf("list empty film roll: %v", err)
	}

	if emptyFrames == nil {
		t.Fatal("expected non-nil empty frame list")
	}

	if len(emptyFrames) != 0 {
		t.Fatalf("expected empty frame list, got %d frames", len(emptyFrames))
	}

	_, err = repository.Create(
		t.Context(),
		otherUser.ID,
		roll.ID,
		nil,
		nil,
	)
	if !errors.Is(err, frame.ErrFilmRollNotFound) {
		t.Fatalf("expected ErrFilmRollNotFound, got %v", err)
	}

	var frameCount int
	if err := pool.QueryRow(
		t.Context(),
		"SELECT COUNT(*) FROM frames WHERE film_roll_id = $1",
		roll.ID,
	).Scan(&frameCount); err != nil {
		t.Fatalf("count frames: %v", err)
	}

	if frameCount != 2 {
		t.Fatalf("expected 2 frames, got %d", frameCount)
	}

	err = repository.Delete(
		t.Context(),
		otherUser.ID,
		first.ID,
	)
	if !errors.Is(err, frame.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if _, err := repository.GetByID(
		t.Context(),
		owner.ID,
		first.ID,
	); err != nil {
		t.Fatalf("frame must remain after forbidden delete: %v", err)
	}

	if err := repository.Delete(
		t.Context(),
		owner.ID,
		first.ID,
	); err != nil {
		t.Fatalf("delete owner's frame: %v", err)
	}

	_, err = repository.GetByID(
		t.Context(),
		owner.ID,
		first.ID,
	)
	if !errors.Is(err, frame.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}

	remainingFrames, err := repository.List(
		t.Context(),
		owner.ID,
		roll.ID,
	)
	if err != nil {
		t.Fatalf("list frames after delete: %v", err)
	}

	if len(remainingFrames) != 1 {
		t.Fatalf("expected 1 frame after delete, got %d", len(remainingFrames))
	}

	if remainingFrames[0].FrameIndex != 2 {
		t.Fatalf(
			"expected remaining frame index 2, got %d",
			remainingFrames[0].FrameIndex,
		)
	}
}
