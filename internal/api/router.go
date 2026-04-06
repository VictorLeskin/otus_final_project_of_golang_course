package api

import (
	"net/http"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/bucket"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/storage"
)

// API объединяет все зависимости для обработчиков.
type API struct {
	bucketManager *bucket.BucketManager
	storage       storage.IPListStorage
}

// NewAPI создает новый экземпляр API.
func NewAPI(bm *bucket.BucketManager, st storage.IPListStorage) *API {
	return &API{
		bucketManager: bm,
		storage:       st,
	}
}

// Router возвращает HTTP маршрутизатор.
func (a *API) Router() http.Handler {
	mux := http.NewServeMux()

	// Основные эндпоинты.
	mux.HandleFunc("POST /check", a.checkHandler)
	mux.HandleFunc("POST /reset", a.resetHandler)

	// Белый список.
	mux.HandleFunc("POST /whitelist/add", a.whitelistAddHandler)
	mux.HandleFunc("POST /whitelist/remove", a.whitelistRemoveHandler)
	mux.HandleFunc("GET /whitelist", a.whitelistHandler)

	// Черный список.
	mux.HandleFunc("POST /blacklist/add", a.blacklistAddHandler)
	mux.HandleFunc("POST /blacklist/remove", a.blacklistRemoveHandler)
	mux.HandleFunc("GET /blacklist", a.blacklistHandler)

	// Статистика (опционально).
	mux.HandleFunc("GET /stats", a.statsHandler)

	return mux
}
