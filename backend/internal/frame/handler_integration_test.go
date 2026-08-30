package frame_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/filmroll"
	"github.com/C-A1R/exponia/backend/internal/frame"
	"github.com/C-A1R/exponia/backend/internal/identity"
	"github.com/C-A1R/exponia/backend/internal/testutil"
	"github.com/C-A1R/exponia/backend/internal/user"
)

func TestFrameHandlers(t *testing.T) {
	pool := testutil.StartPostgres(t)

	userRepository := user.NewRepository(pool)
	owner, err := userRepository.FindOrCreate(
		t.Context(),
		"frame-handler-owner@example.com",
		"Frame Handler Owner",
		"https://auth.example.com",
		"frame-handler-owner",
	)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}

	otherUser, err := userRepository.FindOrCreate(
		t.Context(),
		"frame-handler-other@example.com",
		"Frame Handler Other User",
		"https://auth.example.com",
		"frame-handler-other-user",
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
	service := frame.NewService(repository)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := frame.NewHandler(service, logger)

	mux := http.NewServeMux()
	frame.RegisterRoutes(mux, handler)

	path := "/api/v1/film-rolls/" + strconv.FormatInt(roll.ID, 10) + "/frames"

	request := httptest.NewRequest(
		http.MethodPost,
		path,
		bytes.NewBufferString(`{
			"frame_label": "00",
			"note": "First frame"
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

	var created frame.Frame
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if created.FrameIndex != 1 {
		t.Fatalf("expected frame index 1, got %d", created.FrameIndex)
	}

	if created.FrameLabel == nil || *created.FrameLabel != "00" {
		t.Fatalf("expected frame label %q, got %v", "00", created.FrameLabel)
	}

	getPath := "/api/v1/frames/" + strconv.FormatInt(created.ID, 10)
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

	var fetched frame.Frame
	if err := json.NewDecoder(getResponse.Body).Decode(&fetched); err != nil {
		t.Fatalf("decode fetched frame: %v", err)
	}

	if fetched.ID != created.ID {
		t.Fatalf("expected frame id %d, got %d", created.ID, fetched.ID)
	}

	foreignGetRequest := httptest.NewRequest(http.MethodGet, getPath, nil)
	foreignGetRequest = foreignGetRequest.WithContext(
		identity.WithUserID(foreignGetRequest.Context(), otherUser.ID),
	)

	foreignGetResponse := httptest.NewRecorder()
	mux.ServeHTTP(foreignGetResponse, foreignGetRequest)

	if foreignGetResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d for foreign frame, got %d",
			http.StatusNotFound,
			foreignGetResponse.Code,
		)
	}

	missingGetRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/frames/999999",
		nil,
	)
	missingGetRequest = missingGetRequest.WithContext(
		identity.WithUserID(missingGetRequest.Context(), owner.ID),
	)

	missingGetResponse := httptest.NewRecorder()
	mux.ServeHTTP(missingGetResponse, missingGetRequest)

	if missingGetResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d for missing frame, got %d",
			http.StatusNotFound,
			missingGetResponse.Code,
		)
	}

	updateRequest := httptest.NewRequest(
		http.MethodPut,
		getPath,
		bytes.NewBufferString(`{
			"frame_label": "00A",
			"note": "Updated first frame"
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

	var updated frame.Frame
	if err := json.NewDecoder(updateResponse.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated frame: %v", err)
	}

	if updated.FrameIndex != created.FrameIndex {
		t.Fatalf(
			"expected frame index to remain %d, got %d",
			created.FrameIndex,
			updated.FrameIndex,
		)
	}

	if updated.FrameLabel == nil || *updated.FrameLabel != "00A" {
		t.Fatalf("expected updated frame label %q, got %v", "00A", updated.FrameLabel)
	}

	if updated.Note == nil || *updated.Note != "Updated first frame" {
		t.Fatalf("expected updated note, got %v", updated.Note)
	}

	clearRequest := httptest.NewRequest(
		http.MethodPut,
		getPath,
		bytes.NewBufferString(`{
			"frame_label": null,
			"note": null
		}`),
	)
	clearRequest = clearRequest.WithContext(
		identity.WithUserID(clearRequest.Context(), owner.ID),
	)

	clearResponse := httptest.NewRecorder()
	mux.ServeHTTP(clearResponse, clearRequest)

	if clearResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected clear status %d, got %d: %s",
			http.StatusOK,
			clearResponse.Code,
			clearResponse.Body.String(),
		)
	}

	var cleared frame.Frame
	if err := json.NewDecoder(clearResponse.Body).Decode(&cleared); err != nil {
		t.Fatalf("decode cleared frame: %v", err)
	}

	if cleared.FrameLabel != nil || cleared.Note != nil {
		t.Fatalf(
			"expected cleared fields, got label %v and note %v",
			cleared.FrameLabel,
			cleared.Note,
		)
	}

	foreignUpdateRequest := httptest.NewRequest(
		http.MethodPut,
		getPath,
		bytes.NewBufferString(`{
			"frame_label": "foreign",
			"note": "foreign"
		}`),
	)
	foreignUpdateRequest = foreignUpdateRequest.WithContext(
		identity.WithUserID(foreignUpdateRequest.Context(), otherUser.ID),
	)

	foreignUpdateResponse := httptest.NewRecorder()
	mux.ServeHTTP(foreignUpdateResponse, foreignUpdateRequest)

	if foreignUpdateResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d for foreign update, got %d",
			http.StatusNotFound,
			foreignUpdateResponse.Code,
		)
	}

	unchanged, err := repository.GetByID(t.Context(), owner.ID, created.ID)
	if err != nil {
		t.Fatalf("get frame after foreign update: %v", err)
	}

	if unchanged.FrameLabel != nil || unchanged.Note != nil {
		t.Fatalf(
			"expected frame to remain cleared, got label %v and note %v",
			unchanged.FrameLabel,
			unchanged.Note,
		)
	}

	unauthenticatedRequest := httptest.NewRequest(
		http.MethodPost,
		path,
		bytes.NewBufferString(`{}`),
	)
	unauthenticatedResponse := httptest.NewRecorder()
	mux.ServeHTTP(unauthenticatedResponse, unauthenticatedRequest)

	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d without user, got %d",
			http.StatusUnauthorized,
			unauthenticatedResponse.Code,
		)
	}

	forbiddenRequest := httptest.NewRequest(
		http.MethodPost,
		path,
		bytes.NewBufferString(`{}`),
	)
	forbiddenRequest = forbiddenRequest.WithContext(
		identity.WithUserID(forbiddenRequest.Context(), otherUser.ID),
	)

	forbiddenResponse := httptest.NewRecorder()
	mux.ServeHTTP(forbiddenResponse, forbiddenRequest)

	if forbiddenResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d for another user's film roll, got %d",
			http.StatusNotFound,
			forbiddenResponse.Code,
		)
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

	var listedFrames []frame.Frame
	if err := json.NewDecoder(listResponse.Body).Decode(&listedFrames); err != nil {
		t.Fatalf("decode frame list: %v", err)
	}

	if len(listedFrames) != 2 {
		t.Fatalf("expected 2 listed frames, got %d", len(listedFrames))
	}

	if listedFrames[0].FrameIndex != 1 || listedFrames[1].FrameIndex != 2 {
		t.Fatalf(
			"expected frame indexes [1, 2], got [%d, %d]",
			listedFrames[0].FrameIndex,
			listedFrames[1].FrameIndex,
		)
	}

	if listedFrames[1].ID != second.ID {
		t.Fatalf(
			"expected second frame id %d, got %d",
			second.ID,
			listedFrames[1].ID,
		)
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

	emptyListPath := "/api/v1/film-rolls/" +
		strconv.FormatInt(emptyRoll.ID, 10) +
		"/frames"
	emptyListRequest := httptest.NewRequest(http.MethodGet, emptyListPath, nil)
	emptyListRequest = emptyListRequest.WithContext(
		identity.WithUserID(emptyListRequest.Context(), owner.ID),
	)

	emptyListResponse := httptest.NewRecorder()
	mux.ServeHTTP(emptyListResponse, emptyListRequest)

	if emptyListResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected empty list status %d, got %d",
			http.StatusOK,
			emptyListResponse.Code,
		)
	}

	var emptyFrames []frame.Frame
	if err := json.NewDecoder(emptyListResponse.Body).Decode(&emptyFrames); err != nil {
		t.Fatalf("decode empty frame list: %v", err)
	}

	if emptyFrames == nil || len(emptyFrames) != 0 {
		t.Fatalf("expected empty JSON array, got %v", emptyFrames)
	}

	foreignListRequest := httptest.NewRequest(http.MethodGet, path, nil)
	foreignListRequest = foreignListRequest.WithContext(
		identity.WithUserID(foreignListRequest.Context(), otherUser.ID),
	)

	foreignListResponse := httptest.NewRecorder()
	mux.ServeHTTP(foreignListResponse, foreignListRequest)

	if foreignListResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d for foreign frame list, got %d",
			http.StatusNotFound,
			foreignListResponse.Code,
		)
	}

	missingListRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/film-rolls/999999/frames",
		nil,
	)
	missingListRequest = missingListRequest.WithContext(
		identity.WithUserID(missingListRequest.Context(), owner.ID),
	)

	missingListResponse := httptest.NewRecorder()
	mux.ServeHTTP(missingListResponse, missingListRequest)

	if missingListResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d for missing film roll, got %d",
			http.StatusNotFound,
			missingListResponse.Code,
		)
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

	foreignDeleteRequest := httptest.NewRequest(http.MethodDelete, getPath, nil)
	foreignDeleteRequest = foreignDeleteRequest.WithContext(
		identity.WithUserID(foreignDeleteRequest.Context(), otherUser.ID),
	)

	foreignDeleteResponse := httptest.NewRecorder()
	mux.ServeHTTP(foreignDeleteResponse, foreignDeleteRequest)

	if foreignDeleteResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d for foreign delete, got %d",
			http.StatusNotFound,
			foreignDeleteResponse.Code,
		)
	}

	if _, err := repository.GetByID(
		t.Context(),
		owner.ID,
		created.ID,
	); err != nil {
		t.Fatalf("frame must remain after foreign delete: %v", err)
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
		t.Fatalf("expected empty delete response, got %q", deleteResponse.Body.String())
	}

	deletedGetRequest := httptest.NewRequest(http.MethodGet, getPath, nil)
	deletedGetRequest = deletedGetRequest.WithContext(
		identity.WithUserID(deletedGetRequest.Context(), owner.ID),
	)

	deletedGetResponse := httptest.NewRecorder()
	mux.ServeHTTP(deletedGetResponse, deletedGetRequest)

	if deletedGetResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d after delete, got %d",
			http.StatusNotFound,
			deletedGetResponse.Code,
		)
	}

	remainingFrames, err := repository.List(t.Context(), owner.ID, roll.ID)
	if err != nil {
		t.Fatalf("list frames after delete: %v", err)
	}

	if len(remainingFrames) != 1 {
		t.Fatalf("expected 1 frame after delete, got %d", len(remainingFrames))
	}

	if remainingFrames[0].ID != second.ID || remainingFrames[0].FrameIndex != 2 {
		t.Fatalf(
			"expected remaining frame id %d with index 2, got id %d with index %d",
			second.ID,
			remainingFrames[0].ID,
			remainingFrames[0].FrameIndex,
		)
	}
}
