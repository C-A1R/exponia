package exposure_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/C-A1R/exponia/backend/internal/camera"
	"github.com/C-A1R/exponia/backend/internal/exposure"
	"github.com/C-A1R/exponia/backend/internal/filmroll"
	"github.com/C-A1R/exponia/backend/internal/frame"
	"github.com/C-A1R/exponia/backend/internal/identity"
	"github.com/C-A1R/exponia/backend/internal/lens"
	"github.com/C-A1R/exponia/backend/internal/testutil"
	"github.com/C-A1R/exponia/backend/internal/user"
)

func TestExposureHandlers(t *testing.T) {
	pool := testutil.StartPostgres(t)

	userRepository := user.NewRepository(pool)
	owner, err := userRepository.FindOrCreate(
		t.Context(),
		"exposure-handler-owner@example.com",
		"Exposure Handler Owner",
		"https://auth.example.com",
		"exposure-handler-owner",
	)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}

	otherUser, err := userRepository.FindOrCreate(
		t.Context(),
		"exposure-handler-other@example.com",
		"Exposure Handler Other User",
		"https://auth.example.com",
		"exposure-handler-other-user",
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
	createdCamera, err := cameraRepository.CreateCamera(
		t.Context(),
		owner.ID,
		"Nikon",
		"FM2n",
	)
	if err != nil {
		t.Fatalf("create camera: %v", err)
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
	createdLens, err := lensRepository.Create(
		t.Context(),
		owner.ID,
		"Nikon",
		"Nikkor 50mm f/1.8",
		50,
		1.8,
	)
	if err != nil {
		t.Fatalf("create lens: %v", err)
	}

	otherUsersLens, err := lensRepository.Create(
		t.Context(),
		otherUser.ID,
		"Canon",
		"FD 50mm f/1.8",
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
		&createdCamera.ID,
	)
	if err != nil {
		t.Fatalf("set film roll camera: %v", err)
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
	service := exposure.NewService(repository)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := exposure.NewHandler(service, logger)

	mux := http.NewServeMux()
	exposure.RegisterRoutes(mux, handler)

	path := "/api/v1/frames/" + strconv.FormatInt(createdFrame.ID, 10) + "/exposures"
	request := httptest.NewRequest(
		http.MethodPost,
		path,
		bytes.NewBufferString(`{
			"lens_id": `+strconv.FormatInt(createdLens.ID, 10)+`,
			"aperture": 5.6,
			"shutter_speed_us": 8000,
			"shot_at": "2026-09-01T12:30:00Z",
			"note": "First exposure"
		}`),
	)
	request = request.WithContext(
		identity.WithUserID(request.Context(), owner.ID),
	)

	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	var created exposure.Exposure
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if created.FrameID != createdFrame.ID {
		t.Fatalf("expected frame id %d, got %d", createdFrame.ID, created.FrameID)
	}

	if created.ExposureIndex != 1 {
		t.Fatalf("expected exposure index 1, got %d", created.ExposureIndex)
	}

	if created.CameraID != createdCamera.ID {
		t.Fatalf("expected camera id %d, got %d", createdCamera.ID, created.CameraID)
	}

	if created.LensID == nil || *created.LensID != createdLens.ID {
		t.Fatalf("expected lens id %d, got %v", createdLens.ID, created.LensID)
	}

	if created.Aperture == nil || *created.Aperture != 5.6 {
		t.Fatalf("expected aperture 5.6, got %v", created.Aperture)
	}

	if created.ShutterSpeedUS == nil || *created.ShutterSpeedUS != 8000 {
		t.Fatalf("expected shutter speed 8000, got %v", created.ShutterSpeedUS)
	}

	expectedShotAt := time.Date(2026, time.September, 1, 12, 30, 0, 0, time.UTC)
	if created.ShotAt == nil || !created.ShotAt.Equal(expectedShotAt) {
		t.Fatalf("expected shot time %v, got %v", expectedShotAt, created.ShotAt)
	}

	if created.Note == nil || *created.Note != "First exposure" {
		t.Fatalf("expected note %q, got %v", "First exposure", created.Note)
	}

	stored, err := repository.GetByID(t.Context(), owner.ID, created.ID)
	if err != nil {
		t.Fatalf("get stored exposure: %v", err)
	}

	if stored.ID != created.ID || stored.FrameID != createdFrame.ID {
		t.Fatalf("unexpected stored exposure: %+v", stored)
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

	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid frame id",
			path:       "/api/v1/frames/not-a-number/exposures",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			path:       path,
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown JSON field",
			path:       path,
			body:       `{"unknown": true}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing frame",
			path:       "/api/v1/frames/999999/exposures",
			body:       `{}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid aperture",
			path:       path,
			body:       `{"aperture": 0}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "camera not selected",
			path: "/api/v1/frames/" +
				strconv.FormatInt(frameWithoutCamera.ID, 10) +
				"/exposures",
			body:       `{}`,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				tt.path,
				bytes.NewBufferString(tt.body),
			)
			request = request.WithContext(
				identity.WithUserID(request.Context(), owner.ID),
			)

			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d: %s",
					tt.wantStatus,
					response.Code,
					response.Body.String(),
				)
			}
		})
	}

	storedExposures, err := repository.List(t.Context(), owner.ID, createdFrame.ID)
	if err != nil {
		t.Fatalf("list stored exposures: %v", err)
	}

	if len(storedExposures) != 1 {
		t.Fatalf("expected 1 stored exposure, got %d", len(storedExposures))
	}

	listRequest := httptest.NewRequest(http.MethodGet, path, nil)
	listRequest = listRequest.WithContext(
		identity.WithUserID(listRequest.Context(), owner.ID),
	)

	listResponse := httptest.NewRecorder()
	mux.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected list status %d, got %d: %s",
			http.StatusOK,
			listResponse.Code,
			listResponse.Body.String(),
		)
	}

	var listed []exposure.Exposure
	if err := json.NewDecoder(listResponse.Body).Decode(&listed); err != nil {
		t.Fatalf("decode exposure list: %v", err)
	}

	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("expected created exposure in list, got %+v", listed)
	}

	emptyListPath := "/api/v1/frames/" +
		strconv.FormatInt(frameWithoutCamera.ID, 10) +
		"/exposures"
	emptyListRequest := httptest.NewRequest(http.MethodGet, emptyListPath, nil)
	emptyListRequest = emptyListRequest.WithContext(
		identity.WithUserID(emptyListRequest.Context(), owner.ID),
	)

	emptyListResponse := httptest.NewRecorder()
	mux.ServeHTTP(emptyListResponse, emptyListRequest)

	if emptyListResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected empty list status %d, got %d: %s",
			http.StatusOK,
			emptyListResponse.Code,
			emptyListResponse.Body.String(),
		)
	}

	if emptyListResponse.Body.String() != "[]\n" {
		t.Fatalf("expected empty JSON array, got %q", emptyListResponse.Body.String())
	}

	listErrorTests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "invalid frame id",
			path:       "/api/v1/frames/not-a-number/exposures",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing frame",
			path:       "/api/v1/frames/999999/exposures",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range listErrorTests {
		t.Run("list "+tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			request = request.WithContext(
				identity.WithUserID(request.Context(), owner.ID),
			)

			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d: %s",
					tt.wantStatus,
					response.Code,
					response.Body.String(),
				)
			}
		})
	}

	getPath := "/api/v1/exposures/" + strconv.FormatInt(created.ID, 10)
	getRequest := httptest.NewRequest(http.MethodGet, getPath, nil)
	getRequest = getRequest.WithContext(
		identity.WithUserID(getRequest.Context(), owner.ID),
	)

	getResponse := httptest.NewRecorder()
	mux.ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected get status %d, got %d: %s",
			http.StatusOK,
			getResponse.Code,
			getResponse.Body.String(),
		)
	}

	var fetched exposure.Exposure
	if err := json.NewDecoder(getResponse.Body).Decode(&fetched); err != nil {
		t.Fatalf("decode fetched exposure: %v", err)
	}

	if fetched.ID != created.ID || fetched.FrameID != createdFrame.ID {
		t.Fatalf("unexpected fetched exposure: %+v", fetched)
	}

	getErrorTests := []struct {
		name       string
		path       string
		userID     int64
		wantStatus int
	}{
		{
			name:       "invalid exposure id",
			path:       "/api/v1/exposures/not-a-number",
			userID:     owner.ID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing exposure",
			path:       "/api/v1/exposures/999999",
			userID:     owner.ID,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "foreign exposure",
			path:       getPath,
			userID:     otherUser.ID,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range getErrorTests {
		t.Run("get "+tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			request = request.WithContext(
				identity.WithUserID(request.Context(), tt.userID),
			)

			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d: %s",
					tt.wantStatus,
					response.Code,
					response.Body.String(),
				)
			}
		})
	}

	updateRequest := httptest.NewRequest(
		http.MethodPut,
		getPath,
		bytes.NewBufferString(`{
			"camera_id": `+strconv.FormatInt(secondCamera.ID, 10)+`,
			"lens_id": null,
			"aperture": null,
			"shutter_speed_us": null,
			"shot_at": null,
			"note": null
		}`),
	)
	updateRequest = updateRequest.WithContext(
		identity.WithUserID(updateRequest.Context(), owner.ID),
	)

	updateResponse := httptest.NewRecorder()
	mux.ServeHTTP(updateResponse, updateRequest)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected update status %d, got %d: %s",
			http.StatusOK,
			updateResponse.Code,
			updateResponse.Body.String(),
		)
	}

	var updated exposure.Exposure
	if err := json.NewDecoder(updateResponse.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated exposure: %v", err)
	}

	if updated.ID != created.ID ||
		updated.FrameID != created.FrameID ||
		updated.ExposureIndex != created.ExposureIndex {
		t.Fatalf("expected exposure identity to stay unchanged, got %+v", updated)
	}

	if updated.CameraID != secondCamera.ID {
		t.Fatalf("expected camera id %d, got %d", secondCamera.ID, updated.CameraID)
	}

	if updated.LensID != nil ||
		updated.Aperture != nil ||
		updated.ShutterSpeedUS != nil ||
		updated.ShotAt != nil ||
		updated.Note != nil {
		t.Fatalf("expected optional fields to be cleared, got %+v", updated)
	}

	validUpdateBody := `{"camera_id": ` + strconv.FormatInt(secondCamera.ID, 10) + `}`
	updateErrorTests := []struct {
		name       string
		path       string
		body       string
		userID     int64
		wantStatus int
	}{
		{
			name:       "invalid exposure id",
			path:       "/api/v1/exposures/not-a-number",
			body:       validUpdateBody,
			userID:     owner.ID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "camera id missing",
			path:       getPath,
			body:       `{}`,
			userID:     owner.ID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing exposure",
			path:       "/api/v1/exposures/999999",
			body:       validUpdateBody,
			userID:     owner.ID,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "foreign exposure",
			path: getPath,
			body: `{"camera_id": ` +
				strconv.FormatInt(otherUsersCamera.ID, 10) +
				`}`,
			userID:     otherUser.ID,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "foreign camera",
			path: getPath,
			body: `{"camera_id": ` +
				strconv.FormatInt(otherUsersCamera.ID, 10) +
				`}`,
			userID:     owner.ID,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "foreign lens",
			path: getPath,
			body: `{"camera_id": ` +
				strconv.FormatInt(secondCamera.ID, 10) +
				`, "lens_id": ` +
				strconv.FormatInt(otherUsersLens.ID, 10) +
				`}`,
			userID:     owner.ID,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range updateErrorTests {
		t.Run("update "+tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPut,
				tt.path,
				bytes.NewBufferString(tt.body),
			)
			request = request.WithContext(
				identity.WithUserID(request.Context(), tt.userID),
			)

			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d: %s",
					tt.wantStatus,
					response.Code,
					response.Body.String(),
				)
			}
		})
	}

	storedAfterUpdate, err := repository.GetByID(t.Context(), owner.ID, created.ID)
	if err != nil {
		t.Fatalf("get exposure after update errors: %v", err)
	}

	if storedAfterUpdate.CameraID != secondCamera.ID ||
		storedAfterUpdate.LensID != nil ||
		storedAfterUpdate.Aperture != nil ||
		storedAfterUpdate.ShutterSpeedUS != nil ||
		storedAfterUpdate.ShotAt != nil ||
		storedAfterUpdate.Note != nil {
		t.Fatalf("unexpected stored exposure after update errors: %+v", storedAfterUpdate)
	}

	deleteErrorTests := []struct {
		name       string
		path       string
		userID     int64
		wantStatus int
	}{
		{
			name:       "invalid exposure id",
			path:       "/api/v1/exposures/not-a-number",
			userID:     owner.ID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing exposure",
			path:       "/api/v1/exposures/999999",
			userID:     owner.ID,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "foreign exposure",
			path:       getPath,
			userID:     otherUser.ID,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range deleteErrorTests {
		t.Run("delete "+tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			request = request.WithContext(
				identity.WithUserID(request.Context(), tt.userID),
			)

			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d: %s",
					tt.wantStatus,
					response.Code,
					response.Body.String(),
				)
			}
		})
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, getPath, nil)
	deleteRequest = deleteRequest.WithContext(
		identity.WithUserID(deleteRequest.Context(), owner.ID),
	)

	deleteResponse := httptest.NewRecorder()
	mux.ServeHTTP(deleteResponse, deleteRequest)

	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf(
			"expected delete status %d, got %d: %s",
			http.StatusNoContent,
			deleteResponse.Code,
			deleteResponse.Body.String(),
		)
	}

	if deleteResponse.Body.Len() != 0 {
		t.Fatalf("expected empty delete response body, got %q", deleteResponse.Body.String())
	}

	deletedGetRequest := httptest.NewRequest(http.MethodGet, getPath, nil)
	deletedGetRequest = deletedGetRequest.WithContext(
		identity.WithUserID(deletedGetRequest.Context(), owner.ID),
	)

	deletedGetResponse := httptest.NewRecorder()
	mux.ServeHTTP(deletedGetResponse, deletedGetRequest)

	if deletedGetResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected deleted exposure status %d, got %d",
			http.StatusNotFound,
			deletedGetResponse.Code,
		)
	}

	repeatedDeleteRequest := httptest.NewRequest(http.MethodDelete, getPath, nil)
	repeatedDeleteRequest = repeatedDeleteRequest.WithContext(
		identity.WithUserID(repeatedDeleteRequest.Context(), owner.ID),
	)

	repeatedDeleteResponse := httptest.NewRecorder()
	mux.ServeHTTP(repeatedDeleteResponse, repeatedDeleteRequest)

	if repeatedDeleteResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected repeated delete status %d, got %d",
			http.StatusNotFound,
			repeatedDeleteResponse.Code,
		)
	}

	finalListRequest := httptest.NewRequest(http.MethodGet, path, nil)
	finalListRequest = finalListRequest.WithContext(
		identity.WithUserID(finalListRequest.Context(), owner.ID),
	)

	finalListResponse := httptest.NewRecorder()
	mux.ServeHTTP(finalListResponse, finalListRequest)

	if finalListResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected final list status %d, got %d: %s",
			http.StatusOK,
			finalListResponse.Code,
			finalListResponse.Body.String(),
		)
	}

	if finalListResponse.Body.String() != "[]\n" {
		t.Fatalf("expected empty final exposure list, got %q", finalListResponse.Body.String())
	}
}
