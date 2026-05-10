package response

import (
    "encoding/json"
    "net/http"
)

type APIResponse struct {
    Code int `json:"code"`
    Msj string `json:"msj"`
    Data interface{} `json:"data,omitempty"`
    Metadata interface{} `json:"meta-data,omitempty"`
}

func ResponseJSON(w http.ResponseWriter, status int, resp APIResponse) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(resp)
}

func ResponseSuccess(w http.ResponseWriter, data interface{}, metadata interface{}) {
    ResponseJSON(w, http.StatusOK, APIResponse{
        Code: 0,
        Msj: "OK",
        Data: data,
        Metadata: metadata,
    })
}

func ResponseError(w http.ResponseWriter, httpStatus int, code int, message string) {
    ResponseJSON(w, httpStatus, APIResponse{
        Code: code,
        Msj:  message,
    })
}
