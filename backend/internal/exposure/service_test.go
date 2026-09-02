package exposure

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
)

type fakeExposureRepository struct {
	exposure Exposure

	createCalled  bool
	createUserID  int64
	createFrameID int64
	createInput   CreateInput

	updateCalled     bool
	updateUserID     int64
	updateExposureID int64
	updateInput      UpdateInput
}

func (r *fakeExposureRepository) Create(
	_ context.Context,
	userID int64,
	frameID int64,
	input CreateInput,
) (Exposure, error) {
	r.createCalled = true
	r.createUserID = userID
	r.createFrameID = frameID
	r.createInput = input

	return r.exposure, nil
}

func (r *fakeExposureRepository) GetByID(
	_ context.Context,
	_ int64,
	_ int64,
) (Exposure, error) {
	return r.exposure, nil
}

func (r *fakeExposureRepository) List(
	_ context.Context,
	_ int64,
	_ int64,
) ([]Exposure, error) {
	return []Exposure{r.exposure}, nil
}

func (r *fakeExposureRepository) Update(
	_ context.Context,
	userID int64,
	exposureID int64,
	input UpdateInput,
) (Exposure, error) {
	r.updateCalled = true
	r.updateUserID = userID
	r.updateExposureID = exposureID
	r.updateInput = input

	return r.exposure, nil
}

func (r *fakeExposureRepository) Delete(
	_ context.Context,
	_ int64,
	_ int64,
) error {
	return nil
}

func TestServiceCreateRejectsInvalidOptionalValues(t *testing.T) {
	zeroID := int64(0)
	zeroAperture := 0.0
	nanAperture := math.NaN()
	infiniteAperture := math.Inf(1)
	zeroShutterSpeed := int64(0)

	tests := []struct {
		name    string
		input   CreateInput
		wantErr error
	}{
		{
			name:    "invalid lens id",
			input:   CreateInput{LensID: &zeroID},
			wantErr: ErrInvalidLensID,
		},
		{
			name:    "zero aperture",
			input:   CreateInput{Aperture: &zeroAperture},
			wantErr: ErrInvalidAperture,
		},
		{
			name:    "NaN aperture",
			input:   CreateInput{Aperture: &nanAperture},
			wantErr: ErrInvalidAperture,
		},
		{
			name:    "infinite aperture",
			input:   CreateInput{Aperture: &infiniteAperture},
			wantErr: ErrInvalidAperture,
		},
		{
			name:    "zero shutter speed",
			input:   CreateInput{ShutterSpeedUS: &zeroShutterSpeed},
			wantErr: ErrInvalidShutterSpeed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeExposureRepository{}
			service := NewService(repository)

			_, err := service.Create(t.Context(), 1, 2, tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}

			if repository.createCalled {
				t.Fatal("repository Create must not be called for invalid input")
			}
		})
	}
}

func TestServiceCreatePassesValidInputToRepository(t *testing.T) {
	lensID := int64(3)
	aperture := 5.6
	shutterSpeedUS := int64(8000)
	input := CreateInput{
		LensID:         &lensID,
		Aperture:       &aperture,
		ShutterSpeedUS: &shutterSpeedUS,
	}
	expected := Exposure{ID: 4, FrameID: 2, ExposureIndex: 1}
	repository := &fakeExposureRepository{exposure: expected}
	service := NewService(repository)

	actual, err := service.Create(t.Context(), 1, 2, input)
	if err != nil {
		t.Fatalf("create exposure: %v", err)
	}

	if actual != expected {
		t.Fatalf("got %+v, want %+v", actual, expected)
	}

	if !repository.createCalled ||
		repository.createUserID != 1 ||
		repository.createFrameID != 2 ||
		!reflect.DeepEqual(repository.createInput, input) {
		t.Fatalf("unexpected repository Create call: %+v", repository)
	}
}

func TestServiceUpdateRejectsInvalidCameraID(t *testing.T) {
	repository := &fakeExposureRepository{}
	service := NewService(repository)

	_, err := service.Update(
		t.Context(),
		1,
		4,
		UpdateInput{CameraID: 0},
	)
	if !errors.Is(err, ErrInvalidCameraID) {
		t.Fatalf("expected ErrInvalidCameraID, got %v", err)
	}

	if repository.updateCalled {
		t.Fatal("repository Update must not be called for invalid camera id")
	}
}

func TestServiceUpdatePassesValidInputToRepository(t *testing.T) {
	note := "corrected camera"
	input := UpdateInput{
		CameraID: 5,
		Note:     &note,
	}
	expected := Exposure{ID: 4, FrameID: 2, ExposureIndex: 1, CameraID: 5}
	repository := &fakeExposureRepository{exposure: expected}
	service := NewService(repository)

	actual, err := service.Update(t.Context(), 1, 4, input)
	if err != nil {
		t.Fatalf("update exposure: %v", err)
	}

	if actual != expected {
		t.Fatalf("got %+v, want %+v", actual, expected)
	}

	if !repository.updateCalled ||
		repository.updateUserID != 1 ||
		repository.updateExposureID != 4 ||
		!reflect.DeepEqual(repository.updateInput, input) {
		t.Fatalf("unexpected repository Update call: %+v", repository)
	}
}
