package middlewares

import (
	"net"
	"net/http"

	"github.com/talx-hub/malerter/internal/logger"
)

func CheckNetwork(ipNet *net.IPNet, log *logger.ZeroLogger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		checkFn := func(w http.ResponseWriter, r *http.Request) {
			if ipNet != nil {
				agentIP := net.ParseIP(r.Header.Get("X-Real-IP"))
				if agentIP == nil {
					log.Error().Msg("agent ip is not set or parsing failed. forbidden")
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
				if !ipNet.Contains(agentIP) {
					log.Error().Msg("agent ip is in wrong subnet. forbidden")
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(checkFn)
	}
}
