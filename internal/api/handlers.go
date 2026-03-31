package api

import (
	"encoding/json"
	"net/http"
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

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// checkHandler проверяет авторизацию
func (a *API) checkHandler(w http.ResponseWriter, r *http.Request) {
	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Валидация
	if req.Login == "" || req.Password == "" || req.IP == "" {
		sendError(w, http.StatusBadRequest, "login, password and ip are required")
		return
	}

	// Проверяем через bucketManager
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

// getWhitelistHandler возвращает белый список
func (a *API) getWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	/*
		subnets, err := a.storage.GetAll(r.Context(), models.White)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to get whitelist")
			return
		}
		sendJSON(w, http.StatusOK, ListResponse{Subnets: subnets})
	*/
	panic("Not implemented")
}

// addToWhitelistHandler добавляет подсеть в белый список
func (a *API) addToWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	/*
		var req SubnetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Subnet == "" {
			sendError(w, http.StatusBadRequest, "subnet is required")
			return
		}

		err := a.storage.Add(r.Context(), models.White, req.Subnet)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to add to whitelist")
			return
		}
		sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	*/
	panic("Not implemented")
}

// removeFromWhitelistHandler удаляет подсеть из белого списка
func (a *API) removeFromWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	/*
		var req SubnetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Subnet == "" {
			sendError(w, http.StatusBadRequest, "subnet is required")
			return
		}

		err := a.storage.Remove(r.Context(), models.White, req.Subnet)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to remove from whitelist")
			return
		}
		sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	*/
	panic("Not implemented")
}

// getBlacklistHandler возвращает черный список
func (a *API) getBlacklistHandler(w http.ResponseWriter, r *http.Request) {
	/*
		subnets, err := a.storage.GetAll(r.Context(), models.Black)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to get blacklist")
			return
		}
		sendJSON(w, http.StatusOK, ListResponse{Subnets: subnets})
	*/
	panic("Not implemented")
}

// addToBlacklistHandler добавляет подсеть в черный список
func (a *API) addToBlacklistHandler(w http.ResponseWriter, r *http.Request) {
	/*
		var req SubnetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Subnet == "" {
			sendError(w, http.StatusBadRequest, "subnet is required")
			return
		}

		err := a.storage.Add(r.Context(), models.Black, req.Subnet)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to add to blacklist")
			return
		}
		sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	*/
	panic("Not implemented")
}

// removeFromBlacklistHandler удаляет подсеть из черного списка
func (a *API) removeFromBlacklistHandler(w http.ResponseWriter, r *http.Request) {
	/*
		var req SubnetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Subnet == "" {
			sendError(w, http.StatusBadRequest, "subnet is required")
			return
		}

		err := a.storage.Remove(r.Context(), models.Black, req.Subnet)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to remove from blacklist")
			return
		}
		sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	*/
	panic("Not implemented")
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
