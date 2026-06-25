CREATE TABLE IF NOT EXISTS `servers_202402291556` (
  `token` CHAR (100) PRIMARY KEY UNIQUE NOT NULL,
  `wid` VARCHAR (255) UNIQUE NOT NULL,
  `verified` BOOLEAN NOT NULL DEFAULT FALSE,
  `devel` BOOLEAN NOT NULL DEFAULT FALSE,
  `groups` BOOLEAN DEFAULT NULL,
  `broadcasts` BOOLEAN DEFAULT NULL,
  `readreceipts` BOOLEAN DEFAULT NULL,
  `rejectcalls` BOOLEAN DEFAULT NULL,
  `user` CHAR (255) DEFAULT NULL REFERENCES `users`(`username`),
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO `servers_202402291556`(`token`, `wid`, `verified`, `devel`, `groups`, `broadcasts`, `user`)
SELECT `token`, `wid`, `verified`, `devel`, `handlegroups`, `handlebroadcast`, `user`
FROM `servers`;

DROP TABLE `servers`;
ALTER TABLE `servers_202402291556` RENAME TO `servers`;