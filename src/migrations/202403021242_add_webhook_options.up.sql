CREATE TABLE IF NOT EXISTS `webhooks_202403021242` (
  `context` CHAR (100) NOT NULL REFERENCES `servers`(`token`),
  `url` VARCHAR (255) NOT NULL,
  `forwardinternal` BOOLEAN NOT NULL DEFAULT FALSE,
  `trackid` VARCHAR (100) NOT NULL DEFAULT '',
  `readreceipts` BOOLEAN DEFAULT NULL,
  `groups` BOOLEAN DEFAULT NULL,
  `broadcasts` BOOLEAN DEFAULT NULL,
  `extra` BLOB DEFAULT NULL,
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT `webhooks_pkey` PRIMARY KEY (`context`, `url`)
);

INSERT INTO `webhooks_202403021242` (`context`, `url`,`forwardinternal`,`trackid`,`extra`)
SELECT `context`, `url`, `forwardinternal`, `trackid`, `extra` 
FROM `webhooks`;

DROP TABLE `webhooks`;
ALTER TABLE `webhooks_202403021242` RENAME TO `webhooks`;