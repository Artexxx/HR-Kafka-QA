package api

import (
	"context"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

func RecoveryMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if rvr := recover(); rvr != nil {
				stack := debug.Stack()

				log.Error().
					Interface("panic", rvr).
					Str("method", string(ctx.Method())).
					Str("url", string(ctx.URI().String())).
					Str("remote_addr", ctx.RemoteAddr().String()).
					Str("stack_trace", string(stack)).
					Msg("Recovered from panic")

				pc, file, line, ok := runtime.Caller(3)
				if ok {
					log.Logger.Error().
						Str("file", file).
						Int("line", line).
						Str("function", runtime.FuncForPC(pc).Name()).
						Msg("Panic occurred here")
				}
				ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)

			}
		}()

		next(ctx)
	}
}

// LoggingMiddleware логирует каждый запрос с включением request_id.
func LoggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		requestID, ok := ctx.UserValue("request-id").(string)
		if !ok {
			requestID = uuid.New().String()
			ctx.SetUserValue("request-id", requestID)
		}
		traceCtx, ok := ctx.UserValue("traceContext").(context.Context)
		if !ok {
			traceCtx = context.Background()
		}
		ctxWithRequestID := context.WithValue(traceCtx, "request-id", requestID)
		ctx.SetUserValue("traceContext", ctxWithRequestID)

		ctx.SetUserValue("traceContext", ctxWithRequestID)
		begin := time.Now()
		next(ctx)
		end := time.Now()
		log.Logger.Info().
			Str("request_id", requestID).
			Bytes("method", ctx.Method()).
			Str("url", string(ctx.URI().String())).
			Int("status", ctx.Response.StatusCode()).
			Dur("latency", end.Sub(begin)).
			Msg("Completed request")
	}
}

// CORS middleware для обработки заголовков CORS
func CORS(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return
		}

		next(ctx)
	}
}
