package api

import (
	"encoding/json"
	"net/http"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
)

// CheckRequest запрос на проверку авторизации
type CheckRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

// CheckResponse ответ на проверку
type CheckResponse struct {
	OK bool `json:"ok"`
}

// ResetRequest запрос на сброс bucket'ов
type ResetRequest struct {
	Login string `json:"login"`
	IP    string `json:"ip"`
}

// SubnetRequest запрос на добавление/удаление подсети
type SubnetRequest struct {
	Subnet string `json:"subnet"`
}

// ListResponse ответ со списком подсетей
type ListResponse struct {
	Subnets []string `json:"subnets"`
}

// SuccessfulResponse ответ с ошибкой
type SuccessfulResponse struct {
	Status string `json:"status"`
}

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

func CreateIPList(subnet string, isWhite models.ListType) models.IPList {
	return models.IPList{
		Subnet:  subnet,
		IsWhite: isWhite,
	}
}

func (req CheckRequest) validate() bool {
	if req.Login == "" || req.Password == "" || req.IP == "" {
		return false
	}
	return true
}

// checkHandler проверяет авторизацию
func (a *API) checkHandler(w http.ResponseWriter, r *http.Request) {
	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Валидация
	if !req.validate() {
		sendError(w, http.StatusBadRequest, "login, password and ip are required")
		return
	}

	// Проверяем IP по white/black спискам
	authorized, err := a.storage.IsIPAuthorized(r.Context(), req.IP)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to check IP authorization")
		return
	}

	// Если IP в blacklist → false, если в whitelist → true
	if !authorized {
		sendJSON(w, http.StatusOK, CheckResponse{OK: false})
		return
	}

	// Если IP в whitelist или не в списках — проверяем rate limiter
	ok := a.bucketManager.CheckAuth(req.Login, req.Password, req.IP)
	sendJSON(w, http.StatusOK, CheckResponse{OK: ok})
}

// resetHandler сбрасывает bucket'ы для логина и IP
func (a *API) resetHandler(w http.ResponseWriter, r *http.Request) {
	var req ResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Login == "" && req.IP == "" {
		sendError(w, http.StatusBadRequest, "login or ip required")
		return
	}

	a.bucketManager.ResetAll(req.Login, req.IP)
	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// whitelistHandler возвращает белый список
func (a *API) whitelistHandler(w http.ResponseWriter, r *http.Request) {
	subnets, err := a.storage.GetIpList(r.Context(), models.Black)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to get whitelist")
		return
	}
	sendJSON(w, http.StatusOK, makeListResponse(subnets))
}

// whitelistAddHandler добавляет подсеть в белый список
func (a *API) whitelistAddHandler(w http.ResponseWriter, r *http.Request) {
	var req SubnetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Subnet == "" {
		sendError(w, http.StatusBadRequest, "subnet is required")
		return
	}

	err := a.storage.Add(r.Context(), CreateIPList(req.Subnet, models.White))
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to add to whitelist")
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// whitelistRemoveHandler удаляет подсеть из белого списка
func (a *API) whitelistRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var req SubnetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Subnet == "" {
		sendError(w, http.StatusBadRequest, "subnet is required")
		return
	}

	err := a.storage.Remove(r.Context(), CreateIPList(req.Subnet, models.Black))
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to remove from whitelist")
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func makeListResponse(subnets []models.IPList) (ret ListResponse) {
	for _, item := range subnets {
		ret.Subnets = append(ret.Subnets, item.Subnet)
	}
	return ret
}

// blacklistHandler возвращает черный список
func (a *API) blacklistHandler(w http.ResponseWriter, r *http.Request) {
	subnets, err := a.storage.GetIpList(r.Context(), models.Black)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to get blacklist")
		return
	}
	sendJSON(w, http.StatusOK, makeListResponse(subnets))
}

// blacklistAddHandler добавляет подсеть в черный список
func (a *API) blacklistAddHandler(w http.ResponseWriter, r *http.Request) {
	var req SubnetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Subnet == "" {
		sendError(w, http.StatusBadRequest, "subnet is required")
		return
	}

	err := a.storage.Add(r.Context(), CreateIPList(req.Subnet, models.Black))
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to add to blacklist")
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// blacklistRemoveHandler удаляет подсеть из черного списка
func (a *API) blacklistRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var req SubnetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Subnet == "" {
		sendError(w, http.StatusBadRequest, "subnet is required")
		return
	}

	err := a.storage.Remove(r.Context(), CreateIPList(req.Subnet, models.Black))
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to remove from blacklist")
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// statsHandler возвращает статистику по bucket'ам
func (a *API) statsHandler(w http.ResponseWriter, r *http.Request) {
	/*
		stats := a.bucketManager.GetStats()
		sendJSON(w, http.StatusOK, stats)+
	*/
	panic("Not implemented")
}

// Вспомогательные функции

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, ErrorResponse{Error: message})
}
