package bootstrap

import "github.com/rocketsoftware/open-web-launch/gui"

//go:generate go-bindata-assetfs -pkg bootstrap assets/...
type assets struct {
}

func (*assets) Get(name string) ([]byte, error) {
	return Asset(name)
}

func init() {
	gui.Assets = &assets{}
}
