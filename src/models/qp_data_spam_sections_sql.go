package models

import (
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
)

type QpDataSpamSectionsSql struct {
	db *sqlx.DB
}

func (source QpDataSpamSectionsSql) Find(token string) (*QpSpamSection, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, sql.ErrNoRows
	}

	result := &QpSpamSection{}
	err := source.db.Get(result, `
		SELECT token, position, enabled, label, created_at, updated_at
		FROM spam_sections
		WHERE token = ?
	`, token)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (source QpDataSpamSectionsSql) ListAll() ([]*QpSpamSection, error) {
	result := []*QpSpamSection{}
	err := source.db.Select(&result, `
		SELECT token, position, enabled, label, created_at, updated_at
		FROM spam_sections
		ORDER BY position ASC, token ASC
	`)
	return result, err
}

func (source QpDataSpamSectionsSql) Upsert(section *QpSpamSection) error {
	if section == nil {
		return nil
	}

	section.Token = strings.TrimSpace(section.Token)
	section.Label = strings.TrimSpace(section.Label)
	if section.Token == "" {
		return sql.ErrNoRows
	}

	if section.Position <= 0 {
		next, err := source.NextPosition()
		if err != nil {
			return err
		}
		section.Position = next
	}

	_, err := source.db.NamedExec(`
		INSERT INTO spam_sections (token, position, enabled, label, updated_at)
		VALUES (:token, :position, :enabled, :label, CURRENT_TIMESTAMP)
		ON CONFLICT(token) DO UPDATE SET
			position = excluded.position,
			enabled = excluded.enabled,
			label = excluded.label,
			updated_at = CURRENT_TIMESTAMP
	`, section)
	return err
}

func (source QpDataSpamSectionsSql) UpdatePosition(token string, position int) error {
	_, err := source.db.Exec(`
		UPDATE spam_sections
		SET position = ?, updated_at = CURRENT_TIMESTAMP
		WHERE token = ?
	`, position, strings.TrimSpace(token))
	return err
}

func (source QpDataSpamSectionsSql) Delete(token string) (bool, error) {
	result, err := source.db.Exec("DELETE FROM spam_sections WHERE token = ?", strings.TrimSpace(token))
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (source QpDataSpamSectionsSql) NextPosition() (int, error) {
	var position sql.NullInt64
	if err := source.db.Get(&position, "SELECT MAX(position) FROM spam_sections"); err != nil {
		return 10, err
	}
	if !position.Valid {
		return 10, nil
	}
	return int(position.Int64) + 10, nil
}
