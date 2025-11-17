package httphandlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"pr-reviewer-assigment-service/internal/domain"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeBadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, errorResponse{
		Error: errorBody{
			Code:    string(domain.ErrorNotFound),
			Message: message,
		},
	})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET, POST")
	writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
		Error: errorBody{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "method not allowed",
		},
	})
}

func writeDomainError(w http.ResponseWriter, err error) {
	var dErr *domain.Error
	if errors.As(err, &dErr) {
		log.Printf("%v: %v", dErr.Code, dErr.Message)
		switch dErr.Code {
		case domain.ErrorNotFound:
			writeJSON(w, http.StatusNotFound, errorResponse{
				Error: errorBody{
					Code:    string(dErr.Code),
					Message: dErr.Message,
				},
			})
			return
		case domain.ErrorTeamExists,
			domain.ErrorPRExists,
			domain.ErrorNotAssigned,
			domain.ErrorNoCandidate,
			domain.ErrorPRMerged:
			writeJSON(w, http.StatusConflict, errorResponse{
				Error: errorBody{
					Code:    string(dErr.Code),
					Message: dErr.Message,
				},
			})
			return
		default:
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: errorBody{
					Code:    string(dErr.Code),
					Message: dErr.Message,
				},
			})
			return
		}
	}

	writeJSON(w, http.StatusInternalServerError, errorResponse{
		Error: errorBody{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		},
	})
}
