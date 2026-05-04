package image

import "errors"

var ErrImageNotFound = errors.New("image not found")
var ErrImageURLNotFound = errors.New("image url not found")
var ErrImageExpired = errors.New("image has expired")
var ErrImageStorageNotFound = errors.New("storage record not found for image")
var ErrImageFileTooLarge = errors.New("file size exceeds plan limit")
var ErrImageStorageFull = errors.New("storage quota exceeded")
var ErrImageLimitReached = errors.New("image count limit reached for your plan")
var ErrImageMIMENotAllowed = errors.New("file type not allowed on your plan")
