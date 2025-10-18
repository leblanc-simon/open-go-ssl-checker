package middleware

import (
	"net/http"
	"strings"
)

func LinkMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", strings.Join([]string{
			"</static/css/font.css>; rel=preload; as=style",
			"</static/css/main.css>; rel=preload; as=style",
			"</static/img/logo.png>; rel=preload; as=image",
			"</static/fonts/splash-v7-latin-regular.woff2>; rel=preload; as=font; type=font/woff2; crossorigin",
			"</static/fonts/nunito-v31-latin-regular.woff2>; rel=preload; as=font; type=font/woff2; crossorigin",
			"</static/fonts/nunito-v31-latin-italic.woff2>; rel=preload; as=font; type=font/woff2; crossorigin",
			"</static/fonts/nunito-v31-latin-700.woff2>; rel=preload; as=font; type=font/woff2; crossorigin",
			"</static/fonts/nunito-v31-latin-700italic.woff2>; rel=preload; as=font; type=font/woff2; crossorigin",
		}, ", "))
		next.ServeHTTP(w, r)
	})
}
