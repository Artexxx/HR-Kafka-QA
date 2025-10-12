package api

import "github.com/valyala/fasthttp"

// @Summary Проверка здоровья сервиса
// @Tags    Admin
// @Success 200 {object} okResponse
// @Router  /health [get]
func (s *Service) healthHandler(ctx *fasthttp.RequestCtx) {
	ok(ctx, "OK")
}

// @Summary Полная очистка данных тренажёра (truncate lab.*)
// @Tags    Admin
// @Success 200 {object} okResponse
// @Failure 500 {object} errorResponse
// @Router  /admin/reset [post]
func (s *Service) resetHandler(ctx *fasthttp.RequestCtx) {
	if err := s.events.ResetAll(ctx); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "Все данные очищены")
}
