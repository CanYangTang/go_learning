package middleware

import (
	"log"
	"runtime/debug"

	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/CanYangTang/go_learning/pkg/response"
	"github.com/gin-gonic/gin"
)

// Recovery turns a panic in a downstream handler into a 500 response that uses
// the same error envelope as every other error path.
//
// Without it, a panic escapes the whole handler chain: net/http recovers per
// connection, so the process survives, but the client gets a dropped connection
// instead of a 500 and Logging never gets to write its line.
//
// Mount it inside Logging (after it), so that once the panic is recovered the
// control flow returns into Logging's post-Next() code and the request still
// gets an access log line. Mounting it outside Logging returns a 500 but leaves
// no access log line at all.
//
//	router.Use(RequestID(), Logging(), Recovery(), CORS(allowedOrigins))
//
// handler.writeError is unexported and in another package, so the envelope is
// rebuilt here from pkg/response - the shape stays shared, the helper does not.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			// The panic value can carry hostnames, paths or SQL fragments. It is
			// also the only debugging handle, so it goes to the log - keyed by
			// request ID so the access log line can be matched - never to the client.
			log.Printf("request_id=%s panic=%v\n%s", RequestIDFromContext(c), recovered, debug.Stack())

			appErr := apperror.Internal("internal server error")
			c.AbortWithStatusJSON(appErr.StatusCode, response.ErrorBody{
				Error: response.ErrorPayload{Code: appErr.Code, Message: appErr.Message},
			})
		}()

		c.Next()
	}
}
