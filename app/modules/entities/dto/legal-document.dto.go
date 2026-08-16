package entitiesdto

type UpsertLegalDocument struct {
	Key     string `json:"key" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}
