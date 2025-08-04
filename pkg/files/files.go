package files

import (
	"context"

	"gorm.io/gorm"

	"github.com/COTBU/sotbi.lib/pkg/commonqueries"
)

func GetOriginalFileName(ctx context.Context, db *gorm.DB, file string) (string, error) {
	var attachment struct {
		File             string
		OriginalFileName string
	}

	err := db.
		Debug().
		WithContext(ctx).
		Table(commonqueries.NamedTable(commonqueries.FileLinks, "file_links")).
		Where(`"file" = ?`, file).
		Order("file").
		Limit(1).
		Find(&attachment).
		Error
	if err != nil {
		return "", err
	}

	return attachment.OriginalFileName, nil
}
