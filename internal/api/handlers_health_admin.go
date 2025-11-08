package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
)

type resetRequest struct {
	Password string `json:"password"` // пароль
}

// @Summary Проверка здоровья сервиса
// @Tags    Admin
// @Success 200 {object} okResponse
// @Router  /health [get]
func (s *Service) healthHandler(ctx *fasthttp.RequestCtx) {
	ok(ctx, "OK")
}

// @Summary Полная очистка данных тренажёра (truncate tables.*)
// @Tags    Admin
// @Param   request body resetRequest true "Пароль"
// @Success 200 {object} okResponse
// @Failure 500 {object} errorResponse
// @Router  /admin/reset [post]
func (s *Service) resetHandler(ctx *fasthttp.RequestCtx) {
	var req resetRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, errors.New("required field 'admin_password'"))
		return
	}

	if req.Password != "admin89213password" {
		writeError(ctx, fasthttp.StatusUnauthorized, errors.New("invalid admin password"))
		return
	}

	if err := s.events.ResetAll(ctx); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("events.ResetAll: %w", err))
		return
	}

	ok(ctx, "Все данные очищены")
}
