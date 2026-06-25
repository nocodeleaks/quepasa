CREATE TABLE IF NOT EXISTS `servers_202605121200` (
  `token` CHAR (100) PRIMARY KEY UNIQUE NOT NULL,
  `wid` VARCHAR (255) UNIQUE,
  `verified` BOOLEAN NOT NULL DEFAULT FALSE,
  `devel` BOOLEAN NOT NULL DEFAULT FALSE,
  `metadata` TEXT DEFAULT NULL,
  `groups` INT(1) NOT NULL DEFAULT 0,
  `broadcasts` INT(1) NOT NULL DEFAULT 0,
  `readreceipts` INT(1) NOT NULL DEFAULT 0,
  `calls` INT(1) NOT NULL DEFAULT 0,
  `readupdate` INT(1) NOT NULL DEFAULT 0,
  `direct` INT(1) NOT NULL DEFAULT 0,
  `user` CHAR (255) DEFAULT NULL REFERENCES `users`(`username`),
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO `servers_202605121200`(
  `token`, `wid`, `verified`, `devel`, `metadata`,
  `groups`, `broadcasts`, `readreceipts`, `calls`, `readupdate`, `user`, `timestamp`
)
SELECT
  `token`, `wid`, `verified`, `devel`, `metadata`,
  `groups`, `broadcasts`, `readreceipts`, `calls`, `readupdate`, `user`, `timestamp`
FROM `servers`;

DROP TABLE `servers`;
ALTER TABLE `servers_202605121200` RENAME TO `servers`;
