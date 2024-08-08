package middlewares

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware проверяет, что переданный в заголовке запроса X-Real-IP
// IP-адрес клиента входит в доверенную подсеть
func TrustedSubnetMiddleware(trustedSubnet string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.Header.Get("X-Real-IP")

			if r.URL.Path == "/api/internal/stats" {

				if trustedSubnet == "" || clientIP == "" {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				_, cidr, err := net.ParseCIDR(trustedSubnet)
				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				if !cidr.Contains(net.ParseIP(clientIP)) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
