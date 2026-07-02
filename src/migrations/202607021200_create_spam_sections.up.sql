CREATE TABLE IF NOT EXISTS "spam_sections" (
  "token" CHAR (100) PRIMARY KEY NOT NULL REFERENCES "servers"("token") ON DELETE CASCADE,
  "position" INTEGER NOT NULL DEFAULT 0,
  "enabled" BOOLEAN NOT NULL DEFAULT TRUE,
  "label" TEXT NOT NULL DEFAULT '',
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS "idx_spam_sections_position" ON "spam_sections" ("position", "token");
CREATE INDEX IF NOT EXISTS "idx_spam_sections_enabled" ON "spam_sections" ("enabled");
