-- Rollback: Recreate webhooks table from dispatching
CREATE TABLE IF NOT EXISTS `webhooks` (
  `context` CHAR (100) NOT NULL REFERENCES `servers`(`token`),
  `url` VARCHAR (255) NOT NULL,
  `forwardinternal` BOOLEAN NOT NULL DEFAULT FALSE,
  `trackid` VARCHAR (100) NOT NULL DEFAULT '',
  `readreceipts` INT(1) NOT NULL DEFAULT 0,
  `groups` INT(1) NOT NULL DEFAULT 0,
  `broadcasts` INT(1) NOT NULL DEFAULT 0,
  `extra` BLOB DEFAULT NULL,
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT `webhooks_pkey` PRIMARY KEY (`context`, `url`)
);

-- Migrate webhook data back from dispatching
INSERT INTO `webhooks` (
  `context`, 
  `url`, 
  `forwardinternal`, 
  `trackid`, 
  `readreceipts`, 
  `groups`, 
  `broadcasts`, 
  `extra`, 
  `timestamp`
)
SELECT 
  `context`, 
  `connection_string` as `url`,
  `forwardinternal`, 
  `trackid`, 
  `readreceipts`, 
  `groups`, 
  `broadcasts`, 
  `extra`, 
  `timestamp`
FROM `dispatching`
WHERE `type` = 'webhook';

-- Drop dispatching table
DROP TABLE IF EXISTS `dispatching`;
