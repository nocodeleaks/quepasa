CREATE TABLE IF NOT EXISTS `servers_202602241200` (
  `token` CHAR (100) PRIMARY KEY UNIQUE NOT NULL,
  `wid` VARCHAR (255) UNIQUE,
  `verified` BOOLEAN NOT NULL DEFAULT FALSE,
  `devel` BOOLEAN NOT NULL DEFAULT FALSE,
  `groups` INT(1) NOT NULL DEFAULT 0,
  `broadcasts` INT(1) NOT NULL DEFAULT 0,
  `readreceipts` INT(1) NOT NULL DEFAULT 0,
  `calls` INT(1) NOT NULL DEFAULT 0,
  `readupdate` INT(1) NOT NULL DEFAULT 0,
  `user` CHAR (255) DEFAULT NULL REFERENCES `users`(`username`),
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO `servers_202602241200`(`token`, `wid`, `verified`, `devel`, `groups`, `broadcasts`, `readreceipts`, `calls`, `readupdate`, `user`, `timestamp`)
SELECT `token`,
       CASE WHEN `wid` = '' THEN NULL ELSE `wid` END,
       `verified`, `devel`, `groups`, `broadcasts`, `readreceipts`, `calls`, `readupdate`, `user`, `timestamp`
FROM `servers`;

DROP TABLE `servers`;
ALTER TABLE `servers_202602241200` RENAME TO `servers`;
