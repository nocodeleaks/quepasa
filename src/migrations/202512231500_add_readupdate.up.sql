-- Disable foreign key checks to allow dropping/recreating the servers table
-- Required because dispatching table has FOREIGN KEY referencing servers(token)
PRAGMA foreign_keys=off;

CREATE TABLE IF NOT EXISTS `servers_202512231500` (
  `token` CHAR (100) PRIMARY KEY UNIQUE NOT NULL,
  `wid` VARCHAR (255) UNIQUE NOT NULL,
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

INSERT INTO `servers_202512231500`(`token`, `wid`, `verified`, `devel`, `groups`, `broadcasts`, `readreceipts`, `calls`, `user`)
SELECT `token`, `wid`, `verified`, `devel`, `groups`, `broadcasts`, `readreceipts`, `calls`, `user`
FROM `servers`;

DROP TABLE `servers`;
ALTER TABLE `servers_202512231500` RENAME TO `servers`;

-- Re-enable foreign key checks
PRAGMA foreign_keys=on;
