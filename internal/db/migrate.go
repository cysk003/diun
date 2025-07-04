package db

import (
	"encoding/json"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	bbolterr "go.etcd.io/bbolt/errors"
)

// Migrate runs database migration
func (c *Client) Migrate() error {
	if c.metadata.Version == dbVersion {
		return nil
	}

	migrations := map[int]func(*Client) error{
		2: (*Client).migration2,
	}

	for version := c.metadata.Version + 1; version <= dbVersion; version++ {
		migration, found := migrations[version]
		if !found {
			return errors.Errorf("database migration v%d not found", version)
		}

		log.Info().Msgf("Database migration v%d...", version)
		if err := migration(c); err != nil {
			return errors.Wrapf(err, "database migration v%d failed", version)
		}
	}

	return c.WriteMetadata(Metadata{
		Version: dbVersion,
	})
}

func (c *Client) migration2() error {
	type oldManifest struct {
		Name          string
		Tag           string
		MIMEType      string
		Digest        digest.Digest
		Created       *time.Time
		DockerVersion string
		Labels        map[string]string
		Architecture  string `json:"-"`
		Os            string `json:"-"`
		Layers        []string
	}

	tx, err := c.Begin(true)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, bbolterr.ErrTxClosed) {
			log.Error().Err(err).Msg("Cannot rollback")
		}
	}()

	bucket := tx.Bucket([]byte(bucketManifest))
	curs := bucket.Cursor()
	for k, v := curs.First(); k != nil; k, v = curs.Next() {
		var oldManifest oldManifest
		if err := json.Unmarshal(v, &oldManifest); err != nil {
			return err
		}
		entryBytes, _ := json.Marshal(oldManifest)
		if err := bucket.Put(k, entryBytes); err != nil {
			return err
		}
	}

	return tx.Commit()
}
