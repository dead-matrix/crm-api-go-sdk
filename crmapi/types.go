package crmapi

import "time"

type apiEnvelope struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Code    string          `json:"code"`
	Data    jsonRawEnvelope `json:"data"`
}

type jsonRawEnvelope []byte

type jwtAuthResponse struct {
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expires_at"`
}
