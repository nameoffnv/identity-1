package identity

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

type Identity struct {
	mu        sync.RWMutex
	addresses []string
	keystores map[string]Keystore
}

// New identity service
func New() *Identity {
	s := &Identity{
		keystores: make(map[string]Keystore),
	}
	return s
}

// SetKeystores for handling requests
func (s *Identity) SetKeystores(keystores []Keystore) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, k := range keystores {
		address := fmt.Sprint("0x", k.Address)
		s.addresses = append(s.addresses, address)
		s.keystores[strings.ToLower(address)] = k
	}
}

// Router returns default preset router handler
func (s *Identity) Router() http.Handler {
	r := mux.NewRouter()

	apiRoute := r.PathPrefix("/api/v1").Methods(http.MethodGet).Subrouter()
	apiRoute.Use(apiMiddleware)
	apiRoute.HandleFunc("/info", s.InfoGET)
	apiRoute.HandleFunc("/accounts", s.AccountsGET)
	apiRoute.HandleFunc("/account/{address}", s.AccountAddressGET)

	return r
}

// InfoGET retutrns basic info about server
// GET /api/v1/info
func (s *Identity) InfoGET(w http.ResponseWriter, r *http.Request) {
	info := map[string]string{
		"version":      "1.0",
		"providerName": "Acme",
	}
	writeResponse(info, w, r)
}

// AccountsGET returns list of stored accounts
// GET /api/v1/accounts
func (s *Identity) AccountsGET(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeResponse(s.addresses, w, r)
}

// AccountAddressGET returns full keystore info about account
// GET /api/v1/account/{address}
func (s *Identity) AccountAddressGET(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	address := strings.ToLower(vars["address"])
	if !strings.HasPrefix(address, "0x") {
		http.Error(w, "address must be start with 0x", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	keystore, ok := s.keystores[address]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	io.Copy(w, strings.NewReader(keystore.Keystore))
}

func apiMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func writeResponse(data interface{}, w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
