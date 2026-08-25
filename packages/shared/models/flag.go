// Package models holds the wire/storage types shared across services.
package models

import "time"

// Flag is a feature flag. Key is the stable identifier (Mongo _id) that SDKs
// evaluate against. Enabled gates the flag; Rollout (0-100) is the percentage of
// contexts that get the "on" value when Enabled.
type Flag struct {
	Key         string    `json:"key" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description,omitempty" bson:"description,omitempty"`
	Enabled     bool      `json:"enabled" bson:"enabled"`
	Rollout     int       `json:"rollout" bson:"rollout"`
	Tags        []string  `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

// AuditEntry records a mutation to a flag for the change history.
type AuditEntry struct {
	FlagKey   string    `json:"flag_key" bson:"flag_key"`
	Action    string    `json:"action" bson:"action"` // created|updated|deleted
	Actor     string    `json:"actor" bson:"actor"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}
