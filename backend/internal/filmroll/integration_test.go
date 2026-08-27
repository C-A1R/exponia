package filmroll_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/camera"
	"github.com/C-A1R/exponia/backend/internal/filmroll"
	"github.com/C-A1R/exponia/backend/internal/identity"
	"github.com/C-A1R/exponia/backend/internal/testutil"
	"github.com/C-A1R/exponia/backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupFilmRollAPI(
	t *testing.T,
) (*httptest.Server, *pgxpool.Pool) {
	t.Helper()

	pool := testutil.StartPostgres(t)
	userRepository := user.NewRepository(pool)
	currentUser, err := userRepository.FindOrCreate(
		t.Context(),
		"film-roll-test@example.com",
		"Film Roll Test",
		"https://auth.example.com",
		"film-roll-test-user",
	)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(
		io.Discard,
		nil,
	))

	repository := filmroll.NewRepository(pool)
	service := filmroll.NewService(repository)
	handler := filmroll.NewHandler(service, logger)

	mux := http.NewServeMux()
	filmroll.RegisterRoutes(mux, handler)
	authenticatedHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := identity.WithUserID(r.Context(), currentUser.ID)
			mux.ServeHTTP(w, r.WithContext(ctx))
		},
	)

	return httptest.NewServer(authenticatedHandler), pool
}

func getTestFilmStockAndFormat(
	t *testing.T,
	pool *pgxpool.Pool,
) (int64, int64) {
	t.Helper()

	var filmStockID int64

	err := pool.QueryRow(
		t.Context(),
		`
		SELECT id
		FROM film_stocks
		WHERE manufacturer = 'Kodak'
		  AND name = 'Gold 200'
		`,
	).Scan(&filmStockID)
	if err != nil {
		t.Fatalf("find test film stock: %v", err)
	}

	var formatID int64

	err = pool.QueryRow(
		t.Context(),
		`
		SELECT id
		FROM film_formats
		WHERE code = '35mm'
		`,
	).Scan(&formatID)
	if err != nil {
		t.Fatalf("find test film format: %v", err)
	}

	return filmStockID, formatID
}

func doJSONRequest(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	body any,
) *http.Response {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req, err := http.NewRequest(
		method,
		url,
		bytes.NewReader(data),
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return resp
}

func TestCreateFilmRoll(t *testing.T) {
	server, pool := setupFilmRollAPI(t)
	defer server.Close()

	filmStockID, formatID := getTestFilmStockAndFormat(t, pool)

	resp := doJSONRequest(
		t,
		server.Client(),
		http.MethodPost,
		server.URL+"/api/v1/film-rolls",
		map[string]any{
			"film_stock_id": filmStockID,
			"format_id":     formatID,
		},
	)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			resp.StatusCode,
		)
	}

	var roll filmroll.FilmRoll
	if err := json.NewDecoder(resp.Body).Decode(&roll); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if roll.ID <= 0 {
		t.Fatalf("expected valid id, got %d", roll.ID)
	}

	if roll.FilmStock.Name != "Gold 200" {
		t.Errorf(
			"expected Gold 200, got %q",
			roll.FilmStock.Name,
		)
	}

	if roll.Format.Code != "35mm" {
		t.Errorf(
			"expected 35mm, got %q",
			roll.Format.Code,
		)
	}

	// Главное бизнес-правило:
	// exposure_iso автоматически наследуется от FilmStock.
	if roll.ExposureISO != 200 {
		t.Errorf(
			"expected exposure ISO 200, got %d",
			roll.ExposureISO,
		)
	}

	if roll.Status != filmroll.FilmRollStatusUnused {
		t.Errorf(
			"expected status unused, got %q",
			roll.Status,
		)
	}

	if roll.Camera != nil {
		t.Errorf(
			"expected camera to be nil, got %+v",
			roll.Camera,
		)
	}
}

func TestUpdateFilmRoll(t *testing.T) {
	server, pool := setupFilmRollAPI(t)
	defer server.Close()

	filmStockID, formatID := getTestFilmStockAndFormat(t, pool)

	client := server.Client()

	createResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/api/v1/film-rolls",
		map[string]any{
			"film_stock_id": filmStockID,
			"format_id":     formatID,
		},
	)
	defer createResp.Body.Close()

	var created filmroll.FilmRoll
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created roll: %v", err)
	}

	t.Run("status", func(t *testing.T) {
		resp := doJSONRequest(
			t,
			client,
			http.MethodPatch,
			server.URL+"/api/v1/film-rolls/"+itoa(created.ID)+"/status",
			map[string]any{
				"status": "ready",
			},
		)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var roll filmroll.FilmRoll
		if err := json.NewDecoder(resp.Body).Decode(&roll); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		if roll.Status != filmroll.FilmRollStatusReady {
			t.Errorf("expected ready, got %q", roll.Status)
		}
	})

	t.Run("exposure iso", func(t *testing.T) {
		resp := doJSONRequest(
			t,
			client,
			http.MethodPatch,
			server.URL+"/api/v1/film-rolls/"+itoa(created.ID)+"/exposure-iso",
			map[string]any{
				"exposure_iso": 100,
			},
		)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var roll filmroll.FilmRoll
		if err := json.NewDecoder(resp.Body).Decode(&roll); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		if roll.ExposureISO != 100 {
			t.Errorf(
				"expected exposure ISO 100, got %d",
				roll.ExposureISO,
			)
		}

		// Номинальная чувствительность FilmStock
		// при этом не должна измениться.
		if roll.FilmStock.ISO != 200 {
			t.Errorf(
				"expected film stock ISO 200, got %d",
				roll.FilmStock.ISO,
			)
		}
	})
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}

func TestGetAndListFilmRolls(t *testing.T) {
	server, pool := setupFilmRollAPI(t)
	defer server.Close()

	filmStockID, formatID := getTestFilmStockAndFormat(t, pool)

	client := server.Client()

	createResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/api/v1/film-rolls",
		map[string]any{
			"film_stock_id": filmStockID,
			"format_id":     formatID,
		},
	)

	var created filmroll.FilmRoll
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created roll: %v", err)
	}
	createResp.Body.Close()

	resp, err := client.Get(
		server.URL + "/api/v1/film-rolls/" + itoa(created.ID),
	)
	if err != nil {
		t.Fatalf("get film roll: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var roll filmroll.FilmRoll
	if err := json.NewDecoder(resp.Body).Decode(&roll); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if roll.ID != created.ID {
		t.Errorf(
			"expected id %d, got %d",
			created.ID,
			roll.ID,
		)
	}

	listResp, err := client.Get(
		server.URL + "/api/v1/film-rolls",
	)
	if err != nil {
		t.Fatalf("list film rolls: %v", err)
	}
	defer listResp.Body.Close()

	var rolls []filmroll.FilmRoll
	if err := json.NewDecoder(listResp.Body).Decode(&rolls); err != nil {
		t.Fatalf("decode list: %v", err)
	}

	if len(rolls) == 0 {
		t.Fatal("expected at least one film roll")
	}
}

func TestDeleteFilmRoll(t *testing.T) {
	server, pool := setupFilmRollAPI(t)
	defer server.Close()

	filmStockID, formatID := getTestFilmStockAndFormat(t, pool)

	client := server.Client()

	createResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/api/v1/film-rolls",
		map[string]any{
			"film_stock_id": filmStockID,
			"format_id":     formatID,
		},
	)

	var created filmroll.FilmRoll
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created roll: %v", err)
	}
	createResp.Body.Close()

	req, err := http.NewRequest(
		http.MethodDelete,
		server.URL+"/api/v1/film-rolls/"+itoa(created.ID),
		nil,
	)
	if err != nil {
		t.Fatalf("create delete request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("delete request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf(
			"expected 204, got %d",
			resp.StatusCode,
		)
	}

	getResp, err := client.Get(
		server.URL + "/api/v1/film-rolls/" + itoa(created.ID),
	)
	if err != nil {
		t.Fatalf("get deleted roll: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"expected 404 after delete, got %d",
			getResp.StatusCode,
		)
	}
}

func TestInvalidFilmRollStatus(t *testing.T) {
	server, pool := setupFilmRollAPI(t)
	defer server.Close()

	filmStockID, _ := getTestFilmStockAndFormat(t, pool)

	resp := doJSONRequest(
		t,
		server.Client(),
		http.MethodPatch,
		server.URL+"/api/v1/film-rolls/"+itoa(filmStockID)+"/status",
		map[string]any{
			"status": "banana",
		},
	)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf(
			"expected 400, got %d",
			resp.StatusCode,
		)
	}
}

func TestFilmRollNotFound(t *testing.T) {
	server, _ := setupFilmRollAPI(t)
	defer server.Close()

	resp, err := server.Client().Get(
		server.URL + "/api/v1/film-rolls/999999",
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"expected 404, got %d",
			resp.StatusCode,
		)
	}
}

func TestFilmRollOwnership(t *testing.T) {
	pool := testutil.StartPostgres(t)

	userRepository := user.NewRepository(pool)
	owner, err := userRepository.FindOrCreate(
		t.Context(),
		"film-roll-owner@example.com",
		"Film Roll Owner",
		"https://auth.example.com",
		"film-roll-owner",
	)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}

	otherUser, err := userRepository.FindOrCreate(
		t.Context(),
		"film-roll-other@example.com",
		"Film Roll Other User",
		"https://auth.example.com",
		"film-roll-other-user",
	)
	if err != nil {
		t.Fatalf("create other user: %v", err)
	}

	repository := filmroll.NewRepository(pool)
	service := filmroll.NewService(repository)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := filmroll.NewHandler(service, logger)

	mux := http.NewServeMux()
	filmroll.RegisterRoutes(mux, handler)

	filmStockID, formatID := getTestFilmStockAndFormat(t, pool)
	ownersRoll, err := repository.Create(
		t.Context(),
		owner.ID,
		filmStockID,
		formatID,
	)
	if err != nil {
		t.Fatalf("create owner's film roll: %v", err)
	}

	cameraRepository := camera.NewRepository(pool)
	otherUsersCamera, err := cameraRepository.CreateCamera(
		t.Context(),
		otherUser.ID,
		"Nikon",
		"F3",
	)
	if err != nil {
		t.Fatalf("create other user's camera: %v", err)
	}

	requestAsOtherUser := func(
		method string,
		path string,
		body io.Reader,
	) *httptest.ResponseRecorder {
		t.Helper()

		request := httptest.NewRequest(method, path, body)
		request = request.WithContext(
			identity.WithUserID(request.Context(), otherUser.ID),
		)

		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, request)

		return recorder
	}

	t.Run("cannot get another user's film roll", func(t *testing.T) {
		response := requestAsOtherUser(
			http.MethodGet,
			"/api/v1/film-rolls/"+itoa(ownersRoll.ID),
			nil,
		)

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", response.Code)
		}
	})

	t.Run("cannot list another user's film roll", func(t *testing.T) {
		response := requestAsOtherUser(
			http.MethodGet,
			"/api/v1/film-rolls",
			nil,
		)

		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", response.Code)
		}

		var rolls []filmroll.FilmRoll
		if err := json.NewDecoder(response.Body).Decode(&rolls); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		if len(rolls) != 0 {
			t.Fatalf("expected empty list, got %d film rolls", len(rolls))
		}
	})

	t.Run("cannot update another user's film roll", func(t *testing.T) {
		response := requestAsOtherUser(
			http.MethodPatch,
			"/api/v1/film-rolls/"+itoa(ownersRoll.ID)+"/status",
			bytes.NewBufferString(`{"status":"ready"}`),
		)

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", response.Code)
		}
	})

	t.Run("cannot delete another user's film roll", func(t *testing.T) {
		response := requestAsOtherUser(
			http.MethodDelete,
			"/api/v1/film-rolls/"+itoa(ownersRoll.ID),
			nil,
		)

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", response.Code)
		}
	})

	t.Run("cannot assign another user's camera", func(t *testing.T) {
		request := httptest.NewRequest(
			http.MethodPatch,
			"/api/v1/film-rolls/"+itoa(ownersRoll.ID)+"/camera",
			bytes.NewBufferString(
				`{"camera_id":`+itoa(otherUsersCamera.ID)+`}`,
			),
		)
		request = request.WithContext(
			identity.WithUserID(request.Context(), owner.ID),
		)

		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", response.Code)
		}
	})

	unchangedRoll, err := repository.GetByID(
		t.Context(),
		owner.ID,
		ownersRoll.ID,
	)
	if err != nil {
		t.Fatalf("owner must still access film roll: %v", err)
	}

	if unchangedRoll.Status != filmroll.FilmRollStatusUnused {
		t.Fatalf(
			"expected status %q, got %q",
			filmroll.FilmRollStatusUnused,
			unchangedRoll.Status,
		)
	}

	if unchangedRoll.Camera != nil {
		t.Fatalf("expected camera to remain nil, got %+v", unchangedRoll.Camera)
	}
}
