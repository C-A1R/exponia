package lens_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/identity"
	"github.com/C-A1R/exponia/backend/internal/lens"
	"github.com/C-A1R/exponia/backend/internal/testutil"
	"github.com/C-A1R/exponia/backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func createLensTestUser(
	t *testing.T,
	pool *pgxpool.Pool,
	name string,
) int64 {
	t.Helper()

	repository := user.NewRepository(pool)
	created, err := repository.FindOrCreate(
		t.Context(),
		fmt.Sprintf("%s@example.com", name),
		name,
		"https://auth.example.com",
		name,
	)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return created.ID
}

func lensRequestAsUser(
	request *http.Request,
	userID int64,
) *http.Request {
	return request.WithContext(
		identity.WithUserID(request.Context(), userID),
	)
}

func setupLensAPI(t *testing.T) http.Handler {
	t.Helper()

	pool := testutil.StartPostgres(t)
	userRepository := user.NewRepository(pool)
	currentUser, err := userRepository.FindOrCreate(
		t.Context(),
		"lens-test@example.com",
		"Lens Test",
		"https://auth.example.com",
		"lens-test-user",
	)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	repository := lens.NewRepository(pool)
	service := lens.NewService(repository)
	handler := lens.NewHandler(service, logger)

	mux := http.NewServeMux()
	lens.RegisterRoutes(mux, handler)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := identity.WithUserID(r.Context(), currentUser.ID)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TestListLensesEmpty(t *testing.T) {
	mux := setupLensAPI(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/lenses",
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

func TestLensCRUD(t *testing.T) {
	mux := setupLensAPI(t)

	// CREATE
	createBody := `{
		"manufacturer": "Nikon",
		"model": "Nikkor 50mm f/1.8 Ai-S",
		"focal_length_mm": 50,
		"max_aperture": 1.8
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/lenses",
		strings.NewReader(createBody),
	)

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"create: expected status %d, got %d, body: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var created lens.Lens

	if err := json.NewDecoder(recorder.Body).Decode(&created); err != nil {
		t.Fatalf("create: decode response: %v", err)
	}

	if created.ID == 0 {
		t.Fatal("create: expected non-zero lens id")
	}

	if created.FocalLengthMm != 50 {
		t.Fatalf(
			"create: expected focal length %d, got %d",
			50,
			created.FocalLengthMm,
		)
	}

	if created.MaxAperture != 1.8 {
		t.Fatalf(
			"create: expected max aperture %.1f, got %.1f",
			1.8,
			created.MaxAperture,
		)
	}

	// GET
	request = httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/v1/lenses/%d", created.ID),
		nil,
	)

	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"get: expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var fetched lens.Lens

	if err := json.NewDecoder(recorder.Body).Decode(&fetched); err != nil {
		t.Fatalf("get: decode response: %v", err)
	}

	if fetched.ID != created.ID {
		t.Fatalf(
			"get: expected id %d, got %d",
			created.ID,
			fetched.ID,
		)
	}

	// UPDATE
	updateBody := `{
		"manufacturer": "Nikon",
		"model": "Nikkor 50mm f/1.4 Ai-S",
		"focal_length_mm": 50,
		"max_aperture": 1.4
	}`

	request = httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/api/v1/lenses/%d", created.ID),
		strings.NewReader(updateBody),
	)

	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"update: expected status %d, got %d, body: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var updated lens.Lens

	if err := json.NewDecoder(recorder.Body).Decode(&updated); err != nil {
		t.Fatalf("update: decode response: %v", err)
	}

	expectedModel := "Nikkor 50mm f/1.4 Ai-S"

	if updated.Model != expectedModel {
		t.Fatalf(
			"update: expected model %q, got %q",
			expectedModel,
			updated.Model,
		)
	}

	if updated.MaxAperture != 1.4 {
		t.Fatalf(
			"update: expected max aperture %.1f, got %.1f",
			1.4,
			updated.MaxAperture,
		)
	}

	// DELETE
	request = httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/api/v1/lenses/%d", created.ID),
		nil,
	)

	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"delete: expected status %d, got %d",
			http.StatusNoContent,
			recorder.Code,
		)
	}

	// GET after DELETE
	request = httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/v1/lenses/%d", created.ID),
		nil,
	)

	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"get deleted: expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}

func TestLensOwnership(t *testing.T) {
	pool := testutil.StartPostgres(t)
	ownerID := createLensTestUser(t, pool, "lens-owner")
	otherUserID := createLensTestUser(t, pool, "other-lens-user")

	repository := lens.NewRepository(pool)
	created, err := repository.Create(
		t.Context(),
		ownerID,
		"Nikon",
		"Nikkor 50mm f/1.8 Ai-S",
		50,
		1.8,
	)
	if err != nil {
		t.Fatalf("create owner lens: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := lens.NewService(repository)
	handler := lens.NewHandler(service, logger)
	mux := http.NewServeMux()
	lens.RegisterRoutes(mux, handler)

	listRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/lenses",
		nil,
	)
	listRecorder := httptest.NewRecorder()
	mux.ServeHTTP(
		listRecorder,
		lensRequestAsUser(listRequest, otherUserID),
	)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf(
			"list: expected status %d, got %d",
			http.StatusOK,
			listRecorder.Code,
		)
	}

	if listRecorder.Body.String() != "[]\n" {
		t.Fatalf(
			"list: expected other user to see no lenses, got %q",
			listRecorder.Body.String(),
		)
	}

	getRequest := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/v1/lenses/%d", created.ID),
		nil,
	)
	getRecorder := httptest.NewRecorder()
	mux.ServeHTTP(
		getRecorder,
		lensRequestAsUser(getRequest, otherUserID),
	)

	if getRecorder.Code != http.StatusNotFound {
		t.Fatalf(
			"get: expected status %d, got %d",
			http.StatusNotFound,
			getRecorder.Code,
		)
	}

	updateRequest := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/api/v1/lenses/%d", created.ID),
		strings.NewReader(`{
			"manufacturer": "Leica",
			"model": "Summicron-M 50mm f/2",
			"focal_length_mm": 50,
			"max_aperture": 2
		}`),
	)
	updateRecorder := httptest.NewRecorder()
	mux.ServeHTTP(
		updateRecorder,
		lensRequestAsUser(updateRequest, otherUserID),
	)

	if updateRecorder.Code != http.StatusNotFound {
		t.Fatalf(
			"update: expected status %d, got %d",
			http.StatusNotFound,
			updateRecorder.Code,
		)
	}

	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/api/v1/lenses/%d", created.ID),
		nil,
	)
	deleteRecorder := httptest.NewRecorder()
	mux.ServeHTTP(
		deleteRecorder,
		lensRequestAsUser(deleteRequest, otherUserID),
	)

	if deleteRecorder.Code != http.StatusNotFound {
		t.Fatalf(
			"delete: expected status %d, got %d",
			http.StatusNotFound,
			deleteRecorder.Code,
		)
	}

	ownerLens, err := repository.GetByID(
		t.Context(),
		ownerID,
		created.ID,
	)
	if err != nil {
		t.Fatalf("get owner lens after other user's requests: %v", err)
	}

	if ownerLens.Manufacturer != "Nikon" ||
		ownerLens.Model != "Nikkor 50mm f/1.8 Ai-S" ||
		ownerLens.MaxAperture != 1.8 {
		t.Fatalf(
			"expected owner lens to remain unchanged, got %s %s f/%.1f",
			ownerLens.Manufacturer,
			ownerLens.Model,
			ownerLens.MaxAperture,
		)
	}
}
