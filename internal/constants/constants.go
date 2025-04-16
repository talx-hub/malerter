package constants

import "time"

const (
	KeyContentType     = "Content-Type"
	KeyContentEncoding = "Content-Encoding"
	KeyAcceptEncoding  = "Accept-Encoding"
	KeyHashSHA256      = "HashSHA256"
)

const (
	ContentTypeJSON = "application/json"
	ContentTypeHTML = "text/html"
	ContentTypeText = "text/plain"
)

const (
	PermissionFilePrivate = 0o600
	LogLevelDefault       = "Info"
)

const TimeoutShutdown = 2 * time.Second
const TimeoutAgentRequest = 3 * time.Second

const NoSecret = ""
