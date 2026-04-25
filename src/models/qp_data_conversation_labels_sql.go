package models

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type QpDataConversationLabelSql struct {
	db *sqlx.DB
}

func (source QpDataConversationLabelSql) FindAllForUser(user string, activeOnly *bool) ([]*QpConversationLabel, error) {
	user = strings.TrimSpace(user)
	labels := []*QpConversationLabel{}

	query := "SELECT id, user, name, color, active, timestamp FROM conversation_labels WHERE user = ?"
	args := []any{user}
	if activeOnly != nil {
		query += " AND active = ?"
		args = append(args, *activeOnly)
	}
	query += " ORDER BY name ASC, id ASC"

	if err := source.db.Select(&labels, query, args...); err != nil {
		return nil, err
	}

	return labels, nil
}

func (source QpDataConversationLabelSql) FindByIDForUser(id int64, user string) (*QpConversationLabel, error) {
	label := &QpConversationLabel{}
	err := source.db.Get(label, "SELECT id, user, name, color, active, timestamp FROM conversation_labels WHERE id = ? AND user = ?", id, strings.TrimSpace(user))
	if err != nil {
		return nil, err
	}
	return label, nil
}

func (source QpDataConversationLabelSql) Create(label *QpConversationLabel) (*QpConversationLabel, error) {
	if label == nil {
		return nil, fmt.Errorf("label is required")
	}

	label.Normalize()
	if label.User == "" {
		return nil, fmt.Errorf("user is required")
	}
	if label.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	result, err := source.db.NamedExec(`
		INSERT INTO conversation_labels (user, name, color, active)
		VALUES (:user, :name, :color, :active)
	`, label)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return source.FindByIDForUser(id, label.User)
}

func (source QpDataConversationLabelSql) Update(label *QpConversationLabel) error {
	if label == nil {
		return fmt.Errorf("label is required")
	}

	label.Normalize()
	if label.ID == 0 {
		return fmt.Errorf("id is required")
	}
	if label.User == "" {
		return fmt.Errorf("user is required")
	}
	if label.Name == "" {
		return fmt.Errorf("name is required")
	}

	result, err := source.db.NamedExec(`
		UPDATE conversation_labels
		SET name = :name, color = :color, active = :active
		WHERE id = :id AND user = :user
	`, label)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("label (%d) not found for user %s", label.ID, label.User)
	}

	return nil
}

func (source QpDataConversationLabelSql) Delete(id int64, user string) error {
	user = strings.TrimSpace(user)
	if id == 0 {
		return fmt.Errorf("id is required")
	}
	if user == "" {
		return fmt.Errorf("user is required")
	}

	tx, err := source.db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`
		DELETE FROM conversation_label_links
		WHERE label_id = ? AND label_id IN (
			SELECT id FROM conversation_labels WHERE id = ? AND user = ?
		)
	`, id, id, user); err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM conversation_labels WHERE id = ? AND user = ?", id, user)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("label (%d) not found for user %s", id, user)
	}

	return tx.Commit()
}

func (source QpDataConversationLabelSql) Assign(serverToken string, chatID string, labelID int64, user string) (uint, error) {
	serverToken = strings.TrimSpace(serverToken)
	chatID = strings.TrimSpace(chatID)
	user = strings.TrimSpace(user)

	if serverToken == "" {
		return 0, fmt.Errorf("server token is required")
	}
	if chatID == "" {
		return 0, fmt.Errorf("chatid is required")
	}
	if labelID == 0 {
		return 0, fmt.Errorf("label id is required")
	}

	if _, err := source.FindByIDForUser(labelID, user); err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("label (%d) not found for user %s", labelID, user)
		}
		return 0, err
	}

	result, err := source.db.Exec(`
		INSERT OR IGNORE INTO conversation_label_links (server_token, chat_id, label_id)
		VALUES (?, ?, ?)
	`, serverToken, chatID, labelID)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return uint(affected), nil
}

func (source QpDataConversationLabelSql) Remove(serverToken string, chatID string, labelID int64, user string) (uint, error) {
	serverToken = strings.TrimSpace(serverToken)
	chatID = strings.TrimSpace(chatID)
	user = strings.TrimSpace(user)

	if serverToken == "" {
		return 0, fmt.Errorf("server token is required")
	}
	if chatID == "" {
		return 0, fmt.Errorf("chatid is required")
	}
	if labelID == 0 {
		return 0, fmt.Errorf("label id is required")
	}

	result, err := source.db.Exec(`
		DELETE FROM conversation_label_links
		WHERE server_token = ? AND chat_id = ? AND label_id IN (
			SELECT id FROM conversation_labels WHERE id = ? AND user = ?
		)
	`, serverToken, chatID, labelID, user)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return uint(affected), nil
}

func (source QpDataConversationLabelSql) FindConversationLabels(serverToken string, chatID string, user string) ([]*QpConversationLabel, error) {
	serverToken = strings.TrimSpace(serverToken)
	chatID = strings.TrimSpace(chatID)
	user = strings.TrimSpace(user)

	labels := []*QpConversationLabel{}
	err := source.db.Select(&labels, `
		SELECT cl.id, cl.user, cl.name, cl.color, cl.active, cl.timestamp
		FROM conversation_label_links cll
		INNER JOIN conversation_labels cl ON cl.id = cll.label_id
		WHERE cll.server_token = ? AND cll.chat_id = ? AND cl.user = ?
		ORDER BY cl.name ASC, cl.id ASC
	`, serverToken, chatID, user)
	if err != nil {
		return nil, err
	}

	return labels, nil
}

func (source QpDataConversationLabelSql) FindConversationLabelsMap(serverToken string, user string, chatIDs []string) (map[string][]*QpConversationLabel, error) {
	result := map[string][]*QpConversationLabel{}
	serverToken = strings.TrimSpace(serverToken)
	user = strings.TrimSpace(user)

	if serverToken == "" || user == "" || len(chatIDs) == 0 {
		return result, nil
	}

	normalizedChatIDs := make([]string, 0, len(chatIDs))
	seen := map[string]struct{}{}
	for _, chatID := range chatIDs {
		chatID = strings.TrimSpace(chatID)
		if chatID == "" {
			continue
		}
		if _, ok := seen[chatID]; ok {
			continue
		}
		seen[chatID] = struct{}{}
		normalizedChatIDs = append(normalizedChatIDs, chatID)
	}
	if len(normalizedChatIDs) == 0 {
		return result, nil
	}

	type row struct {
		ChatID string `db:"chat_id"`
		QpConversationLabel
	}

	query, args, err := sqlx.In(`
		SELECT cll.chat_id, cl.id, cl.user, cl.name, cl.color, cl.active, cl.timestamp
		FROM conversation_label_links cll
		INNER JOIN conversation_labels cl ON cl.id = cll.label_id
		WHERE cll.server_token = ? AND cl.user = ? AND cll.chat_id IN (?)
		ORDER BY cll.chat_id ASC, cl.name ASC, cl.id ASC
	`, serverToken, user, normalizedChatIDs)
	if err != nil {
		return nil, err
	}
	query = source.db.Rebind(query)

	rows := []row{}
	if err := source.db.Select(&rows, query, args...); err != nil {
		return nil, err
	}

	for _, item := range rows {
		label := item.QpConversationLabel
		result[item.ChatID] = append(result[item.ChatID], &label)
	}

	return result, nil
}
