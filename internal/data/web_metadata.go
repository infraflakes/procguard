package data

import (
	"database/sql"
)

// WebMetadata holds the cached metadata for a website.
type WebMetadata struct {
	Domain  string `json:"domain"`
	Title   string `json:"title"`
	IconURL string `json:"icon_url"`
}

// GetWebMetadata retrieves the cached metadata for a given domain.
func GetWebMetadata(db *sql.DB, domain string) (*WebMetadata, error) {
	query := "SELECT title, icon_url FROM web_metadata WHERE domain = ?"
	row := db.QueryRow(query, domain)

	var meta WebMetadata
	meta.Domain = domain
	if err := row.Scan(&meta.Title, &meta.IconURL); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No metadata found, not an error
		}
		return nil, err
	}

	return &meta, nil
}
