package api

import (
	"encoding/json"
	"errors"

	"github.com/valyala/fasthttp"
)

var (
	ErrMessageIDRequired = errors.New("поле message_id не передано")

	ErrHistoryIDRequired = errors.New("поле history_id не передано")
	ErrHistoryNotFound   = errors.New("история не найдена")

	ErrEmployeeIDRequired   = errors.New("поле employee id не передано")
	ErrProfileNotFound      = errors.New("профиль сотрудника не найден")
	ErrProfileAlreadyExists = errors.New("профиль сотрудника уже существует")
)

type okResponse struct {
	Status string `json:"status" example:"ok"`
	Msg    string `json:"msg" example:"Готово"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(ctx *fasthttp.RequestCtx, statusCode int, body any) {
	ctx.Response.Header.Set("Content-Type", "application/json; charset=utf-8")
	ctx.SetStatusCode(statusCode)

	_ = json.NewEncoder(ctx).Encode(body)
}

func ok(ctx *fasthttp.RequestCtx, msg string) {
	writeJSON(ctx, fasthttp.StatusOK, okResponse{Status: "ok", Msg: msg})
}

func writeError(ctx *fasthttp.RequestCtx, httpStatus int, err error) {
	ctx.Response.Header.Set("Content-Type", "application/json; charset=utf-8")
	ctx.SetStatusCode(httpStatus)
	_ = json.NewEncoder(ctx).Encode(errorResponse{Code: fasthttp.StatusMessage(httpStatus), Message: err.Error()})

	return
}
