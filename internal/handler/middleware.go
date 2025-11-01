package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
)

func CORS() func(http.Handler) http.Handler {
	origins := strings.TrimSpace(os.Getenv("CORS_ORIGINS"))
	if origins == "" {
		c := cors.AllowAll() //если в CORS_ORIGINS пусто, тогда все разрешаем. В деве удобно, чтоб не париться, в проде обязательно стирать надо, либо назначить, кому разрешаем

		return func(next http.Handler) http.Handler {
			return c.Handler(next)
		}
	} else {
		c := cors.New(cors.Options{
			AllowedOrigins:   splitCSV(origins),
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}, // какие методы API можно вызывать
			AllowedHeaders:   []string{"Content-Type", "Authorization"},                    // какие заголовки можно присылать
			AllowCredentials: true,                                                         // разрешаем cookie/JWT
			MaxAge:           3600,                                                         // браузер не будет спрашивать снова в течение часа
		})

		return func(next http.Handler) http.Handler {
			return c.Handler(next)
		}
	}
}

// splitCSV — разбивает строку по запятым в список строк.
// Используется, чтобы из переменной CORS_ORIGINS получить []string.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
