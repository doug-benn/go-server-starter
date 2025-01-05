package database

import "time"

type Comment struct {
	ID        string    `json:"id,omitempty"`
	Comment   string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
