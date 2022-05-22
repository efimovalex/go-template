package rest

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const CtxKeyTime = "time"

type StatusRecorder struct {
	http.ResponseWriter
	Status   int
	Count    int64
	Response bytes.Buffer
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}
func (r *StatusRecorder) Write(p []byte) (int, error) {
	r.Count += int64(len(p))
	r.Response.Write(p)
	return r.ResponseWriter.Write(p)
}

// LogRequestMiddleware defines a http middleware logs every requests
func (r *REST) LogRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		recorder := &StatusRecorder{w, 0, 0, bytes.Buffer{}}
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(recorder, req)

		r.logger.Debugw("Request received",
			"Method", req.Method,
			"URL", req.URL.String(),
			"UserAgent", req.UserAgent(),
			"Referrer", req.Referer(),
			"RemoteIP", req.RemoteAddr,
			"RequestData", string(bodyBytes),
			"RequestSize", req.ContentLength,
			"ResponseStatus", recorder.Status,
			"ResponseData", recorder.Response.String(),
			"ResponseSize", recorder.Count,
			"Latency", fmt.Sprintf("%.6fs", time.Since(start).Seconds()),
		)
	})
}

// adds the current time to the time context value
// should be added first to the Middleware chain
func addTimeContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), CtxKeyTime, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
