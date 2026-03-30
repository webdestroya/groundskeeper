package utils

import (
	"os"

	"github.com/spf13/afero"
)

var ForcedBaseFS afero.Fs = nil

func BaseFS() afero.Fs {

	if ForcedBaseFS != nil {
		return ForcedBaseFS
	}

	if wd, err := os.Getwd(); err == nil {
		return afero.NewBasePathFs(afero.NewOsFs(), wd)
	}
	return afero.NewOsFs()
}
