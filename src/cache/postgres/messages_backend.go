package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/nocodeleaks/quepasa/cache"
	"github.com/nocodeleaks/quepasa/library"
)

const schema = `
CREATE TABLE IF NOT EXISTS messages (
  msgkey     VARCHAR(300) NOT NULL,
  wid        VARCHAR(255) NOT NULL DEFAULT '',
  chat_id    VARCHAR(255) NOT NULL DEFAULT '',
  timestamp  BIGINT       NOT NULL DEFAULT 0,
  from_me    BOOLEAN      NOT NULL DEFAULT FALSE,
  payload    JSONB        NOT NULL,
  expires_at TIMESTAMPTZ  NULL,
  updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  PRIMARY KEY (msgkey)
);
CREATE INDEX IF NOT EXISTS idx_messages_wid_chat_ts ON messages (wid, chat_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_messages_expires ON messages (expires_at) WHERE expires_at IS NOT NULL;
`

type MessagesBackend struct {
	db    *sql.DB
	redis *redis.Client
	ctx   context.Context
	hot   time.Duration // max redis TTL (RAM protection); 0 => use record TTL as-is
}

func New(p library.DatabaseParameters, rc *cache.RedisConfig, hot time.Duration) (*MessagesBackend, error) {
	ssl := p.SSL
	if ssl == "" {
		ssl = "disable"
	}
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		p.Host, p.Port, p.DataBase, p.User, p.Password, ssl)
	return NewFromDSN(dsn, rc, hot)
}

func NewFromDSN(dsn string, rc *cache.RedisConfig, hot time.Duration) (*MessagesBackend, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("messages schema: %w", err)
	}
	b := &MessagesBackend{db: db, ctx: context.Background(), hot: hot}
	if rc != nil && rc.Host != "" {
		b.redis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", rc.Host, rc.Port),
			Username: rc.Username,
			Password: rc.Password,
			DB:       int(rc.Database),
		})
		if err := b.redis.Ping(b.ctx).Err(); err != nil {
			b.redis = nil // degrade to pg-only
		}
	}
	return b, nil
}

func (b *MessagesBackend) redisKey(key string) string { return "quepasa:messages:" + key }

func expiryPtr(r cache.MessageRecord) *time.Time {
	if r.ExpiresAt.IsZero() {
		return nil // forever
	}
	t := r.ExpiresAt
	return &t
}

func (b *MessagesBackend) Set(key string, record cache.MessageRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	var wid, chat string
	var ts int64
	var fromMe bool
	if m := record.Message; m != nil {
		wid, chat, ts, fromMe = m.Wid, m.Chat.Id, m.Timestamp.Unix(), m.FromMe
	}
	_, err = b.db.Exec(`
		INSERT INTO messages (msgkey, wid, chat_id, timestamp, from_me, payload, expires_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7, now())
		ON CONFLICT (msgkey) DO UPDATE SET
		  wid=EXCLUDED.wid, chat_id=EXCLUDED.chat_id, timestamp=EXCLUDED.timestamp,
		  from_me=EXCLUDED.from_me, payload=EXCLUDED.payload, expires_at=EXCLUDED.expires_at, updated_at=now()`,
		key, wid, chat, ts, fromMe, data, expiryPtr(record))
	if err != nil {
		return err
	}
	if b.redis != nil {
		ttl := time.Duration(0)
		if !record.ExpiresAt.IsZero() {
			ttl = time.Until(record.ExpiresAt)
			if ttl < 0 {
				ttl = time.Second
			}
		}
		if b.hot > 0 && (ttl == 0 || ttl > b.hot) {
			ttl = b.hot
		}
		_ = b.redis.Set(b.ctx, b.redisKey(key), data, ttl).Err()
	}
	return nil
}

func (b *MessagesBackend) Get(key string) (cache.MessageRecord, bool, error) {
	if b.redis != nil {
		if data, err := b.redis.Get(b.ctx, b.redisKey(key)).Bytes(); err == nil {
			var r cache.MessageRecord
			if json.Unmarshal(data, &r) == nil {
				return r, true, nil
			}
		}
	}
	var data []byte
	err := b.db.QueryRow(`SELECT payload FROM messages WHERE msgkey=$1`, key).Scan(&data)
	if err == sql.ErrNoRows {
		return cache.MessageRecord{}, false, nil
	}
	if err != nil {
		return cache.MessageRecord{}, false, err
	}
	var r cache.MessageRecord
	if err := json.Unmarshal(data, &r); err != nil {
		return cache.MessageRecord{}, false, err
	}
	return r, true, nil
}

func (b *MessagesBackend) Delete(key string) error {
	if b.redis != nil {
		_ = b.redis.Del(b.ctx, b.redisKey(key)).Err()
	}
	_, err := b.db.Exec(`DELETE FROM messages WHERE msgkey=$1`, key)
	return err
}

func (b *MessagesBackend) List() ([]cache.MessageRecordEntry, error) {
	rows, err := b.db.Query(`SELECT msgkey, payload FROM messages ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntries(rows)
}

func scanEntries(rows *sql.Rows) ([]cache.MessageRecordEntry, error) {
	var out []cache.MessageRecordEntry
	for rows.Next() {
		var key string
		var data []byte
		if err := rows.Scan(&key, &data); err != nil {
			return nil, err
		}
		var r cache.MessageRecord
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, err
		}
		out = append(out, cache.MessageRecordEntry{Key: key, Record: r})
	}
	return out, rows.Err()
}

func (b *MessagesBackend) Query(f cache.MessageQuery) ([]cache.MessageRecordEntry, int, error) {
	where := []string{"wid = $1"}
	args := []interface{}{f.Wid}
	if f.KeyPrefix != "" {
		args = append(args, f.KeyPrefix+":%")
		where = append(where, fmt.Sprintf("msgkey LIKE $%d", len(args)))
	}
	if f.ChatID != "" {
		args = append(args, f.ChatID)
		where = append(where, fmt.Sprintf("chat_id = $%d", len(args)))
	}
	if f.SinceTimestamp > 0 {
		args = append(args, f.SinceTimestamp)
		where = append(where, fmt.Sprintf("timestamp >= $%d", len(args)))
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := b.db.QueryRow(`SELECT count(*) FROM messages WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	page, limit := f.Page, f.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	args = append(args, limit, (page-1)*limit)
	q := fmt.Sprintf(`SELECT msgkey, payload FROM messages WHERE %s ORDER BY timestamp DESC, msgkey DESC LIMIT $%d OFFSET $%d`,
		cond, len(args)-1, len(args))
	rows, err := b.db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanEntries(rows)
	return items, total, err
}

// StartCleanup deletes expired rows every interval. Call once after New.
// ponytail: fixed loop; add env knob only if retention granularity demands it.
func (b *MessagesBackend) StartCleanup(interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			_, _ = b.db.Exec(`DELETE FROM messages WHERE expires_at IS NOT NULL AND expires_at < now()`)
		}
	}()
}

func (b *MessagesBackend) Close() error {
	if b.redis != nil {
		_ = b.redis.Close()
	}
	return b.db.Close()
}

var _ cache.MessagesBackend = (*MessagesBackend)(nil)
