package logger

import (
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

var log *zap.Logger = zap.NewNop()

func Initialize(level string) error {

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	log = zl
	return nil
}

func WithLoggingRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		// h.ServeHTTP(&lw, r)

		duration := int64(time.Since(start))

		log.Info("logging requests",
			zap.String("uri", uri),
			zap.String("method", method),
			zap.String("duration", strconv.Itoa(int(duration))),
			zap.String("status", strconv.Itoa(responseData.status)),
			zap.String("size", strconv.Itoa(responseData.size)),
		)

		next.ServeHTTP(&lw, r)
	})

	// logFn := func(w http.ResponseWriter, r *http.Request) {
	// 	responseData := &responseData{
	// 		status: 0,
	// 		size:   0,
	// 	}
	// 	lw := loggingResponseWriter{
	// 		ResponseWriter: w,
	// 		responseData:   responseData,
	// 	}

	// 	start := time.Now()
	// 	uri := r.RequestURI
	// 	method := r.Method

	// 	h.ServeHTTP(&lw, r)

	// 	duration := int64(time.Since(start))

	// 	log.Info("logging requests",
	// 		zap.String("uri", uri),
	// 		zap.String("method", method),
	// 		zap.String("duration", strconv.Itoa(int(duration))),
	// 		zap.String("status", strconv.Itoa(responseData.status)),
	// 		zap.String("size", strconv.Itoa(responseData.size)),
	// 	)
	// }

	// return http.HandlerFunc(logFn)
}

// func MyMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 	  // create new context from `r` request context, and assign key `"user"`
// 	  // to value of `"123"`
// 	  ctx := context.WithValue(r.Context(), "user", "123")

// 	  // call the next handler in the chain, passing the response writer and
// 	  // the updated request object with the new context value.
// 	  //
// 	  // note: context.Context values are nested, so any previously set
// 	  // values will be accessible as well, and the new `"user"` key
// 	  // will be accessible from this point forward.
// 	  next.ServeHTTP(w, r.WithContext(ctx))
// 	})
//   }
