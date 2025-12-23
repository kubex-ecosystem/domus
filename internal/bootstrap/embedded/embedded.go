// Package embedded is a dummy package to hold the embed directive for migration files.
package embedded

import (
	"io/fs"

	gl "github.com/kubex-ecosystem/logz"
)

func EmbedMigrationData() (fs.FS, error) {
	gl.Info("Dummy package. Implemented at: bootstrap package")
	return nil, nil
}
