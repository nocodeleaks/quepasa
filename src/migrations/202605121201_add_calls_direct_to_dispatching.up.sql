CREATE TABLE IF NOT EXISTS `dispatching_202605121201` (
  `context` CHAR (100) NOT NULL REFERENCES `servers`(`token`),
  `connection_string` VARCHAR (255) NOT NULL,
  `type` VARCHAR (50) NOT NULL DEFAULT 'webhook',
  `forwardinternal` BOOLEAN NOT NULL DEFAULT FALSE,
  `trackid` VARCHAR (100) NOT NULL DEFAULT '',
  `readreceipts` INT(1) NOT NULL DEFAULT 0,
  `groups` INT(1) NOT NULL DEFAULT 0,
  `broadcasts` INT(1) NOT NULL DEFAULT 0,
  `calls` INT(1) NOT NULL DEFAULT 0,
  `direct` INT(1) NOT NULL DEFAULT 0,
  `extra` BLOB DEFAULT NULL,
  `failure` TIMESTAMP DEFAULT NULL,
  `success` TIMESTAMP DEFAULT NULL,
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT `dispatching_pkey` PRIMARY KEY (`context`, `connection_string`)
);

INSERT INTO `dispatching_202605121201` (
  `context`, `connection_string`, `type`, `forwardinternal`, `trackid`,
  `readreceipts`, `groups`, `broadcasts`,
  `extra`, `failure`, `success`, `timestamp`
)
SELECT
  `context`, `connection_string`, `type`, `forwardinternal`, `trackid`,
  `readreceipts`, `groups`, `broadcasts`,
  `extra`, `failure`, `success`, `timestamp`
FROM `dispatching`;

DROP TABLE `dispatching`;
ALTER TABLE `dispatching_202605121201` RENAME TO `dispatching`;
