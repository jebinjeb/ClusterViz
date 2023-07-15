package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"clusterviz/api"
	gerrors "clusterviz/internal/pkg/gerror"
	"net/http"
)

func respondWithError(ginCtx *gin.Context, err error, msg string) {
	var statusCode int

	defaultMsg := "Internal server error, please check with Marketplace Admin"

	switch gerrors.GetErrorType(err) { // nolint:exhaustive
	case gerrors.ValidationFailed, gerrors.BadRequest:
		log.Errorf("error while processing request: %v msg :%s", err, msg)

		statusCode = http.StatusBadRequest
		defaultMsg = "Re-verify the provided request"
	case gerrors.NotFound:
		log.Errorf("EmailAccount not present: error : %v msg :%s", err, msg)

		statusCode = http.StatusNotFound
		msg = "Requested resource not found"

	case gerrors.TokenNotFound, gerrors.AuthenticationFailed:
		log.Errorf("Connection issue : error : %v msg :%s", err, msg)

		statusCode = http.StatusUnauthorized
		msg = "Invalid token/Session, please login"
	default:
		log.Errorf("Internal issue while processing request : %v , msg : %s", err, msg)

		statusCode = http.StatusInternalServerError
	}

	if msg == "" {
		msg = defaultMsg
	}

	ginCtx.JSON(statusCode, api.ErrorModel{
		Message:   msg,
		ErrorCode: fmt.Sprint(statusCode),
	})
}
