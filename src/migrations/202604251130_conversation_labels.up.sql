CREATE TABLE IF NOT EXISTS "conversation_labels" (
	"id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"user" CHAR (255) NOT NULL REFERENCES "users"("username"),
	"name" VARCHAR (100) NOT NULL COLLATE NOCASE,
	"color" VARCHAR (32) DEFAULT '',
	"active" BOOLEAN NOT NULL DEFAULT TRUE,
	"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT "conversation_labels_user_name_key" UNIQUE ("user", "name")
);

CREATE TABLE IF NOT EXISTS "conversation_label_links" (
	"server_token" CHAR (100) NOT NULL REFERENCES "servers"("token"),
	"chat_id" VARCHAR (255) NOT NULL,
	"label_id" INTEGER NOT NULL REFERENCES "conversation_labels"("id"),
	"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT "conversation_label_links_pkey" PRIMARY KEY ("server_token", "chat_id", "label_id")
);

CREATE INDEX IF NOT EXISTS "idx_conversation_label_links_server_chat" ON "conversation_label_links" ("server_token", "chat_id");
