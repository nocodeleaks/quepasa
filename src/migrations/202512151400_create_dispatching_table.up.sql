-- Create new dispatching table with type column (consolidated migration)
-- This replaces the old webhooks table and supports both webhooks and rabbitmq
CREATE TABLE IF NOT EXISTS `dispatching_temp` (
  `context` CHAR (100) NOT NULL REFERENCES `servers`(`token`),
  `connection_string` VARCHAR (255) NOT NULL,
  `type` VARCHAR (50) NOT NULL DEFAULT 'webhook',
  `forwardinternal` BOOLEAN NOT NULL DEFAULT FALSE,
  `trackid` VARCHAR (100) NOT NULL DEFAULT '',
  `readreceipts` INT(1) NOT NULL DEFAULT 0,
  `groups` INT(1) NOT NULL DEFAULT 0,
  `broadcasts` INT(1) NOT NULL DEFAULT 0,
  `extra` BLOB DEFAULT NULL,
  `timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`context`, `connection_string`)
);

-- Migrate existing webhooks data if table exists
INSERT INTO `dispatching_temp` (
  `context`, 
  `connection_string`, 
  `type`,
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
  `url` as `connection_string`,
  'webhook' as `type`,
  `forwardinternal`, 
  `trackid`, 
  `readreceipts`, 
  `groups`, 
  `broadcasts`, 
  `extra`, 
  `timestamp`
FROM `webhooks`
WHERE EXISTS (SELECT name FROM sqlite_master WHERE type='table' AND name='webhooks');

-- Drop old webhooks table if exists
DROP TABLE IF EXISTS `webhooks`;

-- Drop old dispatching table if exists (in case of re-run)
DROP TABLE IF EXISTS `dispatching`;

-- Rename new table to dispatching
ALTER TABLE `dispatching_temp` RENAME TO `dispatching`;
