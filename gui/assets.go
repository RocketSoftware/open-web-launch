package gui

import "errors"

// AssetGetter provide an interface to retrieve binary asssets packaged with the binary.
type AssetGetter interface {
	// Get the asset by name. Returns an error if it does not exist, or if it otherwise
	// runs into a problem retrieving it.
	Get(name string) ([]byte, error)
}

// Assets holds the AssetFetcher which can retrieve assets for the binary
var Assets AssetGetter

// Asset retrieves an asset by name from the provisioned AssetGetter.
func Asset(name string) ([]byte, error) {
	if Assets == nil {
		return nil, errors.New("no assets provided")
	}

	return Assets.Get(name)
}
