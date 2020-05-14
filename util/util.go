package util

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"net/http/httputil"

	"github.com/han-so1omon/concierge.map/data"
)

// Logger is the global logger for this server
var Logger *zap.SugaredLogger

func LogRequest(next http.Handler) http.Handler {
	var errorResponse = data.ErrorResponse{
		Code: http.StatusInternalServerError, Message: "It's not you it's me.",
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		req, err := httputil.DumpRequest(request, true)
		if err != nil {
			Logger.Errorf("%s\n", err)
			returnErrorResponse(response, request, errorResponse)
			return
		}
		Logger.Infof("%q\n", req)
		next.ServeHTTP(response, request)
	})
}

func returnErrorResponse(response http.ResponseWriter, request *http.Request, errorMessage data.ErrorResponse) {
	httpResponse := &data.ErrorResponse{Code: errorMessage.Code, Message: errorMessage.Message}
	jsonResponse, err := json.Marshal(httpResponse)
	if err != nil {
		panic(err)
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(errorMessage.Code)
	response.Write(jsonResponse)
}
