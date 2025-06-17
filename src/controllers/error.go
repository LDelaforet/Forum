package controllers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type ErrorController struct{}

// ErrorHandler gère l'affichage des pages d'erreur
func (ec *ErrorController) ErrorHandler(temp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Récupérer le code d'erreur depuis l'URL
		errorCode := r.URL.Query().Get("code")
		if errorCode == "" {
			errorCode = "404"
		}

		// Récupérer le message d'erreur depuis l'URL
		errorMessage := r.URL.Query().Get("message")
		if errorMessage == "" {
			errorMessage = "Une erreur est survenue"
		}

		// Récupérer les logs
		logs := []string{
			"Erreur " + errorCode + " - URL: " + r.URL.Path,
			"Méthode: " + r.Method,
			"IP: " + r.RemoteAddr,
			"User-Agent: " + r.UserAgent(),
			"Message: " + errorMessage,
		}

		// Ajouter les paramètres de requête s'il y en a
		if len(r.URL.Query()) > 0 {
			logs = append(logs, "Paramètres de requête:")
			for key, values := range r.URL.Query() {
				logs = append(logs, "  "+key+": "+strings.Join(values, ", "))
			}
		}

		// Ajouter les en-têtes de requête
		logs = append(logs, "En-têtes de requête:")
		for key, values := range r.Header {
			logs = append(logs, "  "+key+": "+strings.Join(values, ", "))
		}

		// Rendre le template avec les données
		if err := temp.ExecuteTemplate(w, "404", struct {
			ErrorCode    string
			ErrorMessage string
			Logs         []string
		}{
			ErrorCode:    errorCode,
			ErrorMessage: errorMessage,
			Logs:         logs,
		}); err != nil {
			http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
			return
		}
	}
}

// WithErrorHandler est un middleware qui gère les erreurs
func (ec *ErrorController) WithErrorHandler(temp *template.Template, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Créer un ResponseWriter personnalisé pour capturer le code d'état
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Exécuter le handler suivant
		next(rw, r)

		// Si le code d'état est une erreur, rediriger vers la page d'erreur
		if rw.statusCode >= 400 {
			// Ne pas rediriger si l'en-tête a déjà été envoyé
			if !rw.wroteHeader {
				params := make([]string, 0)
				params = append(params, "code="+strconv.Itoa(rw.statusCode))
				params = append(params, "message="+http.StatusText(rw.statusCode))
				http.Redirect(w, r, "/error?"+strings.Join(params, "&"), http.StatusSeeOther)
			}
		}
	}
}

// responseWriter est un wrapper pour http.ResponseWriter qui capture le code d'état
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

// WriteHeader capture le code d'état
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write capture l'écriture du corps de la réponse
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
