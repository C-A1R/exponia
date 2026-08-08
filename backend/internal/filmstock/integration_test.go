package filmstock_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/filmstock"
	"github.com/C-A1R/exponia/backend/internal/testutil"
)

func setupFilmStockAPI(t *testing.T) *http.ServeMux {
	t.Helper()

	pool := testutil.StartPostgres(t)

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	repository := filmstock.NewRepository(pool)
	service := filmstock.NewService(repository)
	handler := filmstock.NewHandler(service, logger)

	mux := http.NewServeMux()
	filmstock.RegisterRoutes(mux, handler)

	return mux
}

func TestListFilmStocks(t *testing.T) {
	mux := setupFilmStockAPI(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/film-stocks",
		nil,
	)

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d, body: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var stocks []filmstock.FilmStock

	if err := json.NewDecoder(recorder.Body).Decode(&stocks); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(stocks) == 0 {
		t.Fatal("expected seeded film stocks, got empty list")
	}

	var portra *filmstock.FilmStock

	for i := range stocks {
		if stocks[i].Manufacturer == "Kodak" &&
			stocks[i].Name == "Portra 400" {
			portra = &stocks[i]
			break
		}
	}

	if portra == nil {
		t.Fatal("expected Kodak Portra 400 in film stock catalog")
	}

	if portra.ISO != 400 {
		t.Fatalf(
			"expected Portra 400 ISO 400, got %d",
			portra.ISO,
		)
	}

	if portra.ColorType != "color_negative" {
		t.Fatalf(
			"expected color type %q, got %q",
			"color_negative",
			portra.ColorType,
		)
	}

	if !hasFormat(portra.Formats, "35mm") {
		t.Fatal("expected Portra 400 to support 35mm")
	}

	if !hasFormat(portra.Formats, "120") {
		t.Fatal("expected Portra 400 to support 120")
	}
}

func hasFormat(formats []filmstock.FilmFormat, code string) bool {
	for _, format := range formats {
		if format.Code == code {
			return true
		}
	}

	return false
}

func TestGetFilmStockByID(t *testing.T) {
	mux := setupFilmStockAPI(t)

	listRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/film-stocks",
		nil,
	)

	listRecorder := httptest.NewRecorder()
	mux.ServeHTTP(listRecorder, listRequest)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf(
			"list: expected status %d, got %d",
			http.StatusOK,
			listRecorder.Code,
		)
	}

	var stocks []filmstock.FilmStock

	if err := json.NewDecoder(listRecorder.Body).Decode(&stocks); err != nil {
		t.Fatalf("list: decode response: %v", err)
	}

	var stockID int64

	for _, stock := range stocks {
		if stock.Manufacturer == "Kodak" &&
			stock.Name == "Portra 400" {
			stockID = stock.ID
			break
		}
	}

	if stockID == 0 {
		t.Fatal("expected Kodak Portra 400 in catalog")
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/film-stocks/"+strconv.FormatInt(stockID, 10),
		nil,
	)

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d, body: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var stock filmstock.FilmStock

	if err := json.NewDecoder(recorder.Body).Decode(&stock); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if stock.ID != stockID {
		t.Fatalf(
			"expected id %d, got %d",
			stockID,
			stock.ID,
		)
	}

	if stock.Name != "Portra 400" {
		t.Fatalf(
			"expected name %q, got %q",
			"Portra 400",
			stock.Name,
		)
	}

	if !hasFormat(stock.Formats, "35mm") ||
		!hasFormat(stock.Formats, "120") {
		t.Fatalf(
			"expected Portra 400 formats 35mm and 120, got %+v",
			stock.Formats,
		)
	}
}

func TestGetFilmStockNotFound(t *testing.T) {
	mux := setupFilmStockAPI(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/film-stocks/999999",
		nil,
	)

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}
