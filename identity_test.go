package identity_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/endpass/identity"

	"github.com/stretchr/testify/assert"
)

func TestAccountsHandler(t *testing.T) {
	assert := assert.New(t)

	keystores := []identity.Keystore{
		identity.Keystore{Address: "2ad1b3ccb3ec85337ca2dbfa99845c37e06ab238"},
		identity.Keystore{Address: "2ad1b3ccb3ec85337ca2dbfa99845c37e06ab239"},
	}

	is := identity.New()
	router := is.Router()

	// Empty keystore
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)

	router.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal("application/json", res.Header.Get("Content-Type"))

	var response []string
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(err)
	assert.Len(response, 0)

	// With keystore
	is.SetKeystores(keystores)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)

	router.ServeHTTP(w, r)

	res = w.Result()
	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal("application/json", res.Header.Get("Content-Type"))

	err = json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(err)

	assert.Contains(response, "0x2ad1b3ccb3ec85337ca2dbfa99845c37e06ab238")
	assert.Contains(response, "0x2ad1b3ccb3ec85337ca2dbfa99845c37e06ab239")
}

func TestAccountHandler(t *testing.T) {
	assert := assert.New(t)

	keystoresStr := []string{
		`{"address":"2ad1b3ccb3ec85337ca2dbfa99845c37e06ab400","crypto":{"cipher":"aes-128-ctr","cipherparams":{"iv":"f28cd5b3cb26a7f560c234f5cc5c0787"},"ciphertext":"0cbc13b52c07838c2b9ea606870ee651e930962f194bbf98ef860373632d43e0","kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"0b6cd52039d0829315c50e8ec15767c9a82b11d1e22bb200e0eb363ccbdca8d8"},"mac":"35ccb9e2562927b3af92c632a8d6ada241f58e4c5197e37bb0e0ad4782639ec6"},"id":"9cb14b83-6021-45a3-a1a9-4a86387310fc","version":3}`,
		`{"address":"2ad1b3ccb3ec85337ca2dbfa99845c37e06ab401","crypto":{"cipher":"aes-128-ctr","cipherparams":{"iv":"f28cd5b3cb26a7z341c234f5cc5c0530"},"ciphertext":"0cbc13b52c07838c2b9ea606870ee651e930962f194bbf98ef860373632d43e0","kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"0b6cd52039d0829315c50e8ec15767c9a82b11d1e22bb200e0eb363ccbdca8d8"},"mac":"35ccb9e2562927b3af92c632a8d6ada241f58e4c5197e37bb0e0ad4782715ec8"},"id":"9cb14b83-5921-45a3-a1a9-4a86387310fc","version":3}`,
	}

	keystores := make([]identity.Keystore, len(keystoresStr))
	for i, k := range keystoresStr {
		var keystore identity.Keystore

		err := json.Unmarshal([]byte(k), &keystore)
		assert.NoError(err)

		keystore.Keystore = k
		keystores[i] = keystore
	}

	is := identity.New()
	router := is.Router()

	is.SetKeystores(keystores)

	// Not found
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, fmt.Sprint("/api/v1/account/", "0xbadkey"), nil)

	router.ServeHTTP(w, r)

	res := w.Result()
	assert.Equal(http.StatusNotFound, res.StatusCode)

	// Bad request
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, fmt.Sprint("/api/v1/account/", keystores[0].Address), nil)

	router.ServeHTTP(w, r)

	res = w.Result()
	assert.Equal(http.StatusBadRequest, res.StatusCode)

	// Normal
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, fmt.Sprint("/api/v1/account/0x", strings.ToUpper(keystores[1].Address)), nil)

	router.ServeHTTP(w, r)

	res = w.Result()
	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal("application/json", res.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(err)

	assert.Equal(keystores[1].Keystore, string(body))
}
