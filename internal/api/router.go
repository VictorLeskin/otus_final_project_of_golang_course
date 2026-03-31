package api

import (
	"net/http"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/bucket"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/storage"
)

// API объединяет все зависимости для обработчиков
type API struct {
	bucketManager *bucket.BucketManager
	storage       storage.IPListStorage
}

// NewAPI создает новый экземпляр API
func NewAPI(bm *bucket.BucketManager, st storage.IPListStorage) *API {
	return &API{
		bucketManager: bm,
		storage:       st,
	}
}

// Router возвращает HTTP маршрутизатор
func (a *API) Router() http.Handler {
	mux := http.NewServeMux()

	// Основные эндпоинты
	mux.HandleFunc("POST /check", a.checkHandler)
	mux.HandleFunc("POST /reset", a.resetHandler)

	// Белый список
	mux.HandleFunc("POST /whitelist/add", a.addToWhitelistHandler)
	mux.HandleFunc("POST /whitelist/remove", a.removeFromWhitelistHandler)
	mux.HandleFunc("GET /whitelist", a.getWhitelistHandler)

	// Черный список
	mux.HandleFunc("POST /blacklist/add", a.addToBlacklistHandler)
	mux.HandleFunc("POST /blacklist/remove", a.removeFromBlacklistHandler)
	mux.HandleFunc("GET /blacklist", a.getBlacklistHandler)

	// Статистика (опционально)
	mux.HandleFunc("GET /stats", a.statsHandler)

	return mux
}
