package api

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type okResponse struct {
	Status string `json:"status" example:"ok"`
	Msg    string `json:"msg" example:"Готово"`
}

type errorResponse struct {
	Status  string `json:"status" example:"error"`
	Code    string `json:"code" example:"invalid_json"`
	Message string `json:"message" example:"Некорректный JSON"`
}

type listResponse struct {
	Items  any `json:"items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type objectResponse struct {
	Data any `json:"data"`
}

func writeJSON(ctx *fasthttp.RequestCtx, statusCode int, body any) {
	ctx.Response.Header.Set("Content-Type", "application/json; charset=utf-8")
	ctx.SetStatusCode(statusCode)
	_ = json.NewEncoder(ctx).Encode(body)
}

func ok(ctx *fasthttp.RequestCtx, msg string) {
	writeJSON(ctx, fasthttp.StatusOK, okResponse{Status: "ok", Msg: msg})
}
func badRequest(ctx *fasthttp.RequestCtx, code, message string) {
	writeJSON(ctx, fasthttp.StatusBadRequest, errorResponse{
		Status:  "error",
		Code:    code,
		Message: message,
	})
}
func notFound(ctx *fasthttp.RequestCtx, code, message string) {
	writeJSON(ctx, fasthttp.StatusNotFound, errorResponse{
		Status:  "error",
		Code:    code,
		Message: message,
	})
}
func notImplemented(ctx *fasthttp.RequestCtx, code, message string) {
	writeJSON(ctx, fasthttp.StatusNotImplemented, errorResponse{
		Status:  "error",
		Code:    code,
		Message: message,
	})
}
func serverError(ctx *fasthttp.RequestCtx, err error) {
	writeJSON(ctx, fasthttp.StatusInternalServerError, errorResponse{
		Status:  "error",
		Code:    "internal_error",
		Message: "Внутренняя ошибка сервера",
	})
}
