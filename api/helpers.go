package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/validator"
	"github.com/sirupsen/logrus"
)

type ID struct {
	ID int64 `json:"id"`
}
type Message struct {
	Message string `json:"message"`
}

func GetIDParam(c *gin.Context) (int64, error) {
	metricID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, errory.ParseErrors.Builder().Wrap(err).WithMessage("ID parameter cannot be parsed").Create()
	}
	return metricID, validator.ValidateID(metricID)
}

func GetUserContext(c *gin.Context) (*auth.UserContext, error) {
	u, exists := c.Get(middleware.UserContextContextKey)
	if !exists {
		return nil, errory.ProcessingErrors.New("could not extract UserContext")
	}
	userContext, ok := u.(*auth.UserContext)
	if !ok {
		return nil, errory.ProcessingErrors.New("type assertion for UserContext failed")
	}
	return userContext, nil
}

func setErrorResponse(c *gin.Context, err error, logger logrus.FieldLogger) {
	details := errory.GetErrorDetails(err)
	var errorBuilder *logrus.Entry
	var message string
	if details.Stacktrace == "" {
		errorBuilder = logger.WithError(errory.APIErrors.New(details.Message)).WithField("status", details.Status)
		message = "problem handling request"
	} else {
		errorBuilder = logger.WithError(err).WithField("status", details.Status).WithField("stacktrace", details.Stacktrace)
		message = details.LogErrorMessage
	}

	if details.Status == http.StatusInternalServerError {
		errorBuilder.Error(message)
	} else {
		errorBuilder.Warn(message)
	}

	c.JSON(details.Status, Message{Message: details.Message})
}

func setIDResponse(code int, id int64, c *gin.Context) {
	c.JSON(code, ID{ID: id})
}
