package camera

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/testutil"
)

func TestListCamerasEmpty(t *testing.T) {
	pool := testutil.StartPostgres(t)

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	repository := NewRepository(pool)
	service := NewService(repository)
	handler := NewHandler(service, logger)

	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/cameras",
		nil,
	)

	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if recorder.Body.String() != "[]\n" {
		t.Fatalf(
			"expected empty array, got %q",
			recorder.Body.String(),
		)
	}
}

func TestCameraCRUD(t *testing.T) {
	pool := testutil.StartPostgres(t)

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	repository := NewRepository(pool)
	service := NewService(repository)
	handler := NewHandler(service, logger)

	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	// 1. POST
	createBody := `{
		"manufacturer": "Nikon",
		"model": "FM2n"
	}`

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/cameras",
		strings.NewReader(createBody),
	)
	createRequest.Header.Set("Content-Type", "application/json")

	createRecorder := httptest.NewRecorder()

	mux.ServeHTTP(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf(
			"create: expected status %d, got %d, body: %s",
			http.StatusCreated,
			createRecorder.Code,
			createRecorder.Body.String(),
		)
	}

	var created Camera

	if err := json.NewDecoder(createRecorder.Body).Decode(&created); err != nil {
		t.Fatalf("create: decode response: %v", err)
	}

	if created.ID == 0 {
		t.Fatal("create: expected non-zero camera id")
	}

	if created.Manufacturer != "Nikon" {
		t.Fatalf(
			"create: expected manufacturer %q, got %q",
			"Nikon",
			created.Manufacturer,
		)
	}

	if created.Model != "FM2n" {
		t.Fatalf(
			"create: expected model %q, got %q",
			"FM2n",
			created.Model,
		)
	}

	// 2. GET
	getRequest := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/v1/cameras/%d", created.ID),
		nil,
	)

	getRecorder := httptest.NewRecorder()

	mux.ServeHTTP(getRecorder, getRequest)

	if getRecorder.Code != http.StatusOK {
		t.Fatalf(
			"get: expected status %d, got %d, body: %s",
			http.StatusOK,
			getRecorder.Code,
			getRecorder.Body.String(),
		)
	}

	var fetched Camera

	if err := json.NewDecoder(getRecorder.Body).Decode(&fetched); err != nil {
		t.Fatalf("get: decode response: %v", err)
	}

	if fetched.ID != created.ID {
		t.Fatalf(
			"get: expected id %d, got %d",
			created.ID,
			fetched.ID,
		)
	}

	// 3. PUT
	updateBody := `{
		"manufacturer": "Nikon",
		"model": "FM3A"
	}`

	updateRequest := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/api/v1/cameras/%d", created.ID),
		strings.NewReader(updateBody),
	)
	updateRequest.Header.Set("Content-Type", "application/json")

	updateRecorder := httptest.NewRecorder()

	mux.ServeHTTP(updateRecorder, updateRequest)

	if updateRecorder.Code != http.StatusOK {
		t.Fatalf(
			"update: expected status %d, got %d, body: %s",
			http.StatusOK,
			updateRecorder.Code,
			updateRecorder.Body.String(),
		)
	}

	var updated Camera

	if err := json.NewDecoder(updateRecorder.Body).Decode(&updated); err != nil {
		t.Fatalf("update: decode response: %v", err)
	}

	if updated.Model != "FM3A" {
		t.Fatalf(
			"update: expected model %q, got %q",
			"FM3A",
			updated.Model,
		)
	}

	// 4. DELETE
	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/api/v1/cameras/%d", created.ID),
		nil,
	)

	deleteRecorder := httptest.NewRecorder()

	mux.ServeHTTP(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf(
			"delete: expected status %d, got %d, body: %s",
			http.StatusNoContent,
			deleteRecorder.Code,
			deleteRecorder.Body.String(),
		)
	}

	// 5. GET after DELETE
	getDeletedRequest := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/v1/cameras/%d", created.ID),
		nil,
	)

	getDeletedRecorder := httptest.NewRecorder()

	mux.ServeHTTP(getDeletedRecorder, getDeletedRequest)

	if getDeletedRecorder.Code != http.StatusNotFound {
		t.Fatalf(
			"get deleted: expected status %d, got %d, body: %s",
			http.StatusNotFound,
			getDeletedRecorder.Code,
			getDeletedRecorder.Body.String(),
		)
	}
}
