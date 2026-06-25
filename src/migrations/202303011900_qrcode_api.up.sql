CREATE TABLE IF NOT EXISTS `servers` (
  `token` CHAR (100) PRIMARY KEY UNIQUE NOT NULL,
  `wid` VARCHAR (255) UNIQUE NOT NULL,
  `verified` BOOLEAN NOT NULL DEFAULT FALSE,
  `devel` BOOLEAN NOT NULL DEFAULT FALSE,
  `handlegroups` BOOLEAN NOT NULL DEFAULT TRUE,
  `handlebroadcast` BOOLEAN NOT NULL DEFAULT FALSE,
  `user` CHAR (36) DEFAULT NULL REFERENCES `users`(`username`),
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO `servers` (`token`, `wid`, `verified`, `devel`, `handlegroups`, `handlebroadcast`, `user`, `timestamp`)
SELECT `bots`.`token`, `bots`.`id` || "@migrated", `bots`.`is_verified`, `bots`.`devel`, `bots`.`handlegroups`, `bots`.`handlebroadcast`, `users`.`username`, `bots`.`updated_at` FROM `bots` LEFT JOIN `users` ON `bots`.`user_id` = `users`.`id`;

DROP TABLE `bots`;

CREATE TABLE IF NOT EXISTS `users_temp` (
  `username` CHAR (255) PRIMARY KEY NOT NULL,
  `password` VARCHAR (255) NOT NULL,
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO `users_temp` (`username`, `password`)
SELECT `username`, `password` FROM `users`;
DROP TABLE `users`;
ALTER TABLE `users_temp` RENAME TO `users`;


CREATE TABLE IF NOT EXISTS `webhooks_temp` (
  `context` CHAR (255) NOT NULL REFERENCES `servers`(`token`),
  `url` VARCHAR (255) NOT NULL,
  `forwardinternal` BOOLEAN NOT NULL DEFAULT FALSE,
  `trackid` VARCHAR (100) NOT NULL DEFAULT '',
  `extra` BLOB DEFAULT NULL,
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT `webhooks_pkey` PRIMARY KEY (`context`, `url`)
);

INSERT INTO `webhooks_temp` (`context`, `url`,`forwardinternal`,`trackid`,`extra`)
SELECT `context` || "@migrated", `url`, `forwardinternal`,`trackid`,`extra` FROM `webhooks`;
DROP TABLE `webhooks`;
ALTER TABLE `webhooks_temp` RENAME TO `webhooks`;