CREATE TABLE IF NOT EXISTS `servers_202403141920` (
  `token` CHAR (100) PRIMARY KEY UNIQUE NOT NULL,
  `wid` VARCHAR (255) UNIQUE NOT NULL,
  `verified` BOOLEAN NOT NULL DEFAULT FALSE,
  `devel` BOOLEAN NOT NULL DEFAULT FALSE,
  `groups` INT(1) NOT NULL DEFAULT 0,
  `broadcasts` INT(1) NOT NULL DEFAULT 0,
  `readreceipts` INT(1) NOT NULL DEFAULT 0,
  `calls` INT(1) NOT NULL DEFAULT 0,
  `user` CHAR (255) DEFAULT NULL REFERENCES `users`(`username`),
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO `servers_202403141920`(`token`, `wid`, `verified`, `devel`, `groups`, `broadcasts`, `readreceipts`, `calls`, `user`)
SELECT `token`, `wid`, `verified`, `devel`,
CASE WHEN `groups` IS NULL THEN 0 ELSE CASE WHEN `groups` IS true THEN 1 ELSE -1 END END,
CASE WHEN `broadcasts` IS NULL THEN 0 ELSE CASE WHEN `broadcasts` IS true THEN 1 ELSE -1 END END,
CASE WHEN `readreceipts` IS NULL THEN 0 ELSE CASE WHEN `readreceipts` IS true THEN 1 ELSE -1 END END,
CASE WHEN `rejectcalls` IS NULL THEN 0 ELSE CASE WHEN `rejectcalls` IS true THEN -1 ELSE 1 END END,
`user`
FROM `servers`;

DROP TABLE `servers`;
ALTER TABLE `servers_202403141920` RENAME TO `servers`;

CREATE TABLE IF NOT EXISTS `webhooks_202403141920` (
	`context` CHAR (100) NOT NULL REFERENCES `servers`(`token`),
	`url` VARCHAR (255) NOT NULL,
	`forwardinternal` BOOLEAN NOT NULL DEFAULT FALSE,
	`trackid` VARCHAR (100) NOT NULL DEFAULT '',
	`groups` INT(1) NOT NULL DEFAULT 0,
	`broadcasts` INT(1) NOT NULL DEFAULT 0,
	`readreceipts` INT(1) NOT NULL DEFAULT 0,
	`calls` INT(1) NOT NULL DEFAULT 0,
	`extra` BLOB DEFAULT NULL,
	`timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT `webhooks_pkey` PRIMARY KEY (`context`, `url`)
	);

    
INSERT INTO `webhooks_202403141920`(`context`, `url`, `forwardinternal`, `trackid`, `readreceipts`, `groups`, `broadcasts`,  `extra`, `timestamp`)
SELECT `context`, `url`, `forwardinternal`, `trackid`,
CASE WHEN `readreceipts` IS NULL THEN 0 ELSE CASE WHEN `readreceipts` IS true THEN 1 ELSE -1 END END,
CASE WHEN `groups` IS NULL THEN 0 ELSE CASE WHEN `groups` IS true THEN 1 ELSE -1 END END,
CASE WHEN `broadcasts` IS NULL THEN 0 ELSE CASE WHEN `broadcasts` IS true THEN 1 ELSE -1 END END,
`extra`, `timestamp`
FROM `webhooks`;

DROP TABLE `webhooks`;
ALTER TABLE `webhooks_202403141920` RENAME TO `webhooks`;