package utils

import (
	"net/http"
	"net/url"
)

// RedirectToError redirige vers la page d'erreur avec le code et le message spécifiés
func RedirectToError(w http.ResponseWriter, r *http.Request, code string, message string) {
	// Préparer les paramètres de l'URL
	params := url.Values{}
	params.Add("code", code)
	params.Add("message", message)

	// Rediriger vers la page d'erreur
	http.Redirect(w, r, "/error?"+params.Encode(), http.StatusSeeOther)
}
