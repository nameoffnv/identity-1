package identity

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
)

// Keystore V3
type Keystore struct {
	Address  string `json:"address"`
	Keystore string `json:"-"`
}

// LoadKeystores from directory
func LoadKeystores(p string) ([]Keystore, error) {
	fs, err := ioutil.ReadDir(p)
	if err != nil {
		return nil, err
	}

	keystores := make([]Keystore, len(fs))
	for i, f := range fs {
		if f.IsDir() {
			continue
		}

		fp := path.Join(p, f.Name())
		b, err := ioutil.ReadFile(fp)
		if err != nil {
			return nil, err
		}

		var k Keystore
		if err := json.Unmarshal(b, &k); err != nil {
			return nil, fmt.Errorf("file '%s': %s", fp, err.Error())
		}
		k.Keystore = string(b)

		keystores[i] = k
	}

	return keystores, nil
}
