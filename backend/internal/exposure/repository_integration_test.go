package exposure_test

import (
	"errors"
	"testing"
	"time"

	"github.com/C-A1R/exponia/backend/internal/camera"
	"github.com/C-A1R/exponia/backend/internal/exposure"
	"github.com/C-A1R/exponia/backend/internal/filmroll"
	"github.com/C-A1R/exponia/backend/internal/frame"
	"github.com/C-A1R/exponia/backend/internal/lens"
	"github.com/C-A1R/exponia/backend/internal/testutil"
	"github.com/C-A1R/exponia/backend/internal/user"
)

func TestRepository(t *testing.T) {
	pool := testutil.StartPostgres(t)

	userRepository := user.NewRepository(pool)
	owner, err := userRepository.FindOrCreate(
		t.Context(),
		"exposure-owner@example.com",
		"Exposure Owner",
		"https://auth.example.com",
		"exposure-owner",
	)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}

	otherUser, err := userRepository.FindOrCreate(
		t.Context(),
		"exposure-other@example.com",
		"Exposure Other User",
		"https://auth.example.com",
		"exposure-other-user",
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

	cameraRepository := camera.NewRepository(pool)
	firstCamera, err := cameraRepository.CreateCamera(
		t.Context(),
		owner.ID,
		"Nikon",
		"FM2n",
	)
	if err != nil {
		t.Fatalf("create first camera: %v", err)
	}

	secondCamera, err := cameraRepository.CreateCamera(
		t.Context(),
		owner.ID,
		"Nikon",
		"F3",
	)
	if err != nil {
		t.Fatalf("create second camera: %v", err)
	}

	otherUsersCamera, err := cameraRepository.CreateCamera(
		t.Context(),
		otherUser.ID,
		"Canon",
		"F-1",
	)
	if err != nil {
		t.Fatalf("create other user's camera: %v", err)
	}

	lensRepository := lens.NewRepository(pool)
	ownersLens, err := lensRepository.Create(
		t.Context(),
		owner.ID,
		"Nikon",
		"Nikkor 50mm f/1.8",
		50,
		1.8,
	)
	if err != nil {
		t.Fatalf("create owner's lens: %v", err)
	}

	otherUsersLens, err := lensRepository.Create(
		t.Context(),
		otherUser.ID,
		"Canon",
		"EF 50mm f/1.8",
		50,
		1.8,
	)
	if err != nil {
		t.Fatalf("create other user's lens: %v", err)
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

	roll, err = filmRollRepository.UpdateCamera(
		t.Context(),
		owner.ID,
		roll.ID,
		&firstCamera.ID,
	)
	if err != nil {
		t.Fatalf("set first film roll camera: %v", err)
	}

	frameRepository := frame.NewRepository(pool)
	createdFrame, err := frameRepository.Create(
		t.Context(),
		owner.ID,
		roll.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("create frame: %v", err)
	}

	repository := exposure.NewRepository(pool)
	aperture := 5.6
	shutterSpeedUS := int64(8000)
	shotAt := time.Date(2026, time.September, 1, 12, 30, 0, 0, time.UTC)
	note := "First exposure"

	first, err := repository.Create(
		t.Context(),
		owner.ID,
		createdFrame.ID,
		exposure.CreateInput{
			LensID:         &ownersLens.ID,
			Aperture:       &aperture,
			ShutterSpeedUS: &shutterSpeedUS,
			ShotAt:         &shotAt,
			Note:           &note,
		},
	)
	if err != nil {
		t.Fatalf("create first exposure: %v", err)
	}

	if first.ExposureIndex != 1 {
		t.Fatalf("expected first exposure index 1, got %d", first.ExposureIndex)
	}

	if first.CameraID != firstCamera.ID {
		t.Fatalf("expected camera id %d, got %d", firstCamera.ID, first.CameraID)
	}

	if first.LensID == nil || *first.LensID != ownersLens.ID {
		t.Fatalf("expected lens id %d, got %v", ownersLens.ID, first.LensID)
	}

	if first.Aperture == nil || *first.Aperture != aperture {
		t.Fatalf("expected aperture %.1f, got %v", aperture, first.Aperture)
	}

	if first.ShutterSpeedUS == nil || *first.ShutterSpeedUS != shutterSpeedUS {
		t.Fatalf(
			"expected shutter speed %d, got %v",
			shutterSpeedUS,
			first.ShutterSpeedUS,
		)
	}

	if first.ShotAt == nil || !first.ShotAt.Equal(shotAt) {
		t.Fatalf("expected shot time %v, got %v", shotAt, first.ShotAt)
	}

	if first.Note == nil || *first.Note != note {
		t.Fatalf("expected note %q, got %v", note, first.Note)
	}

	_, err = filmRollRepository.UpdateCamera(
		t.Context(),
		owner.ID,
		roll.ID,
		&secondCamera.ID,
	)
	if err != nil {
		t.Fatalf("set second film roll camera: %v", err)
	}

	second, err := repository.Create(
		t.Context(),
		owner.ID,
		createdFrame.ID,
		exposure.CreateInput{},
	)
	if err != nil {
		t.Fatalf("create second exposure: %v", err)
	}

	if second.ExposureIndex != 2 {
		t.Fatalf("expected second exposure index 2, got %d", second.ExposureIndex)
	}

	if second.CameraID != secondCamera.ID {
		t.Fatalf("expected camera id %d, got %d", secondCamera.ID, second.CameraID)
	}

	if second.LensID != nil ||
		second.Aperture != nil ||
		second.ShutterSpeedUS != nil ||
		second.ShotAt != nil ||
		second.Note != nil {
		t.Fatalf("expected nullable fields to be nil, got %+v", second)
	}

	fetched, err := repository.GetByID(
		t.Context(),
		owner.ID,
		first.ID,
	)
	if err != nil {
		t.Fatalf("get owner's exposure: %v", err)
	}

	if fetched.ID != first.ID {
		t.Fatalf("expected exposure id %d, got %d", first.ID, fetched.ID)
	}

	_, err = repository.GetByID(
		t.Context(),
		otherUser.ID,
		first.ID,
	)
	if !errors.Is(err, exposure.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for foreign exposure, got %v", err)
	}

	exposures, err := repository.List(
		t.Context(),
		owner.ID,
		createdFrame.ID,
	)
	if err != nil {
		t.Fatalf("list owner's exposures: %v", err)
	}

	if len(exposures) != 2 {
		t.Fatalf("expected 2 exposures, got %d", len(exposures))
	}

	if exposures[0].ExposureIndex != 1 || exposures[1].ExposureIndex != 2 {
		t.Fatalf(
			"expected exposure indexes [1, 2], got [%d, %d]",
			exposures[0].ExposureIndex,
			exposures[1].ExposureIndex,
		)
	}

	_, err = repository.List(
		t.Context(),
		otherUser.ID,
		createdFrame.ID,
	)
	if !errors.Is(err, exposure.ErrFrameNotFound) {
		t.Fatalf("expected ErrFrameNotFound for foreign frame, got %v", err)
	}

	_, err = repository.List(
		t.Context(),
		owner.ID,
		999999,
	)
	if !errors.Is(err, exposure.ErrFrameNotFound) {
		t.Fatalf("expected ErrFrameNotFound for missing frame, got %v", err)
	}

	var storedFirstCameraID int64
	if err := pool.QueryRow(
		t.Context(),
		"SELECT camera_id FROM exposures WHERE id = $1",
		first.ID,
	).Scan(&storedFirstCameraID); err != nil {
		t.Fatalf("get stored first exposure camera: %v", err)
	}

	if storedFirstCameraID != firstCamera.ID {
		t.Fatalf(
			"expected first exposure camera id %d, got %d",
			firstCamera.ID,
			storedFirstCameraID,
		)
	}

	updated, err := repository.Update(
		t.Context(),
		owner.ID,
		first.ID,
		exposure.UpdateInput{CameraID: secondCamera.ID},
	)
	if err != nil {
		t.Fatalf("update exposure: %v", err)
	}

	if updated.ID != first.ID ||
		updated.FrameID != first.FrameID ||
		updated.ExposureIndex != first.ExposureIndex {
		t.Fatalf("expected exposure identity to stay unchanged, got %+v", updated)
	}

	if updated.CameraID != secondCamera.ID {
		t.Fatalf("expected updated camera id %d, got %d", secondCamera.ID, updated.CameraID)
	}

	if updated.LensID != nil ||
		updated.Aperture != nil ||
		updated.ShutterSpeedUS != nil ||
		updated.ShotAt != nil ||
		updated.Note != nil {
		t.Fatalf("expected optional fields to be cleared, got %+v", updated)
	}

	storedUpdated, err := repository.GetByID(t.Context(), owner.ID, first.ID)
	if err != nil {
		t.Fatalf("get updated exposure: %v", err)
	}

	if storedUpdated.CameraID != secondCamera.ID ||
		storedUpdated.LensID != nil ||
		storedUpdated.Aperture != nil ||
		storedUpdated.ShutterSpeedUS != nil ||
		storedUpdated.ShotAt != nil ||
		storedUpdated.Note != nil {
		t.Fatalf("expected stored exposure to match update, got %+v", storedUpdated)
	}

	_, err = repository.Update(
		t.Context(),
		owner.ID,
		first.ID,
		exposure.UpdateInput{CameraID: otherUsersCamera.ID},
	)
	if !errors.Is(err, exposure.ErrCameraNotFound) {
		t.Fatalf("expected ErrCameraNotFound for foreign camera, got %v", err)
	}

	_, err = repository.Update(
		t.Context(),
		owner.ID,
		first.ID,
		exposure.UpdateInput{
			CameraID: secondCamera.ID,
			LensID:   &otherUsersLens.ID,
		},
	)
	if !errors.Is(err, exposure.ErrLensNotFound) {
		t.Fatalf("expected ErrLensNotFound for foreign lens, got %v", err)
	}

	_, err = repository.Update(
		t.Context(),
		otherUser.ID,
		first.ID,
		exposure.UpdateInput{CameraID: otherUsersCamera.ID},
	)
	if !errors.Is(err, exposure.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for foreign exposure, got %v", err)
	}

	_, err = repository.Update(
		t.Context(),
		owner.ID,
		999999,
		exposure.UpdateInput{CameraID: secondCamera.ID},
	)
	if !errors.Is(err, exposure.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing exposure, got %v", err)
	}

	_, err = repository.Create(
		t.Context(),
		owner.ID,
		createdFrame.ID,
		exposure.CreateInput{LensID: &otherUsersLens.ID},
	)
	if !errors.Is(err, exposure.ErrLensNotFound) {
		t.Fatalf("expected ErrLensNotFound for foreign lens, got %v", err)
	}

	_, err = repository.Create(
		t.Context(),
		otherUser.ID,
		createdFrame.ID,
		exposure.CreateInput{},
	)
	if !errors.Is(err, exposure.ErrFrameNotFound) {
		t.Fatalf("expected ErrFrameNotFound for foreign frame, got %v", err)
	}

	rollWithoutCamera, err := filmRollRepository.Create(
		t.Context(),
		owner.ID,
		filmStockID,
		formatID,
	)
	if err != nil {
		t.Fatalf("create film roll without camera: %v", err)
	}

	frameWithoutCamera, err := frameRepository.Create(
		t.Context(),
		owner.ID,
		rollWithoutCamera.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("create frame without camera: %v", err)
	}

	emptyExposures, err := repository.List(
		t.Context(),
		owner.ID,
		frameWithoutCamera.ID,
	)
	if err != nil {
		t.Fatalf("list empty frame exposures: %v", err)
	}

	if emptyExposures == nil {
		t.Fatal("expected non-nil empty exposure list")
	}

	if len(emptyExposures) != 0 {
		t.Fatalf("expected empty exposure list, got %d", len(emptyExposures))
	}

	_, err = repository.Create(
		t.Context(),
		owner.ID,
		frameWithoutCamera.ID,
		exposure.CreateInput{},
	)
	if !errors.Is(err, exposure.ErrCameraNotSelected) {
		t.Fatalf("expected ErrCameraNotSelected without camera, got %v", err)
	}

	err = repository.Delete(t.Context(), otherUser.ID, first.ID)
	if !errors.Is(err, exposure.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when deleting foreign exposure, got %v", err)
	}

	err = repository.Delete(t.Context(), owner.ID, first.ID)
	if err != nil {
		t.Fatalf("delete exposure: %v", err)
	}

	_, err = repository.GetByID(t.Context(), owner.ID, first.ID)
	if !errors.Is(err, exposure.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for deleted exposure, got %v", err)
	}

	err = repository.Delete(t.Context(), owner.ID, first.ID)
	if !errors.Is(err, exposure.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when deleting exposure twice, got %v", err)
	}

	remainingExposures, err := repository.List(
		t.Context(),
		owner.ID,
		createdFrame.ID,
	)
	if err != nil {
		t.Fatalf("list exposures after delete: %v", err)
	}

	if len(remainingExposures) != 1 {
		t.Fatalf("expected 1 exposure after delete, got %d", len(remainingExposures))
	}

	if remainingExposures[0].ID != second.ID || remainingExposures[0].ExposureIndex != 2 {
		t.Fatalf("expected second exposure to keep index 2, got %+v", remainingExposures[0])
	}

	var exposureCount int
	if err := pool.QueryRow(
		t.Context(),
		"SELECT COUNT(*) FROM exposures WHERE frame_id = $1",
		createdFrame.ID,
	).Scan(&exposureCount); err != nil {
		t.Fatalf("count exposures: %v", err)
	}

	if exposureCount != 1 {
		t.Fatalf("expected 1 exposure, got %d", exposureCount)
	}
}
