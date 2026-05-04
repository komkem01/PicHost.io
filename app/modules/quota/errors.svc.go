package quota

import (
	"errors"
)

var (
	ErrQuotaFileTooLarge      = errors.New("quota: file size exceeds plan limit")
	ErrQuotaStorageFull       = errors.New("quota: storage quota exceeded")
	ErrQuotaImageLimitReached = errors.New("quota: image count limit reached")
	ErrQuotaMIMENotAllowed    = errors.New("quota: file type not allowed on this plan")
)
