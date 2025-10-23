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

-- Migrate existing webhooks data: only copy rows that have a matching server
-- token to respect the foreign key constraint. Rows referencing missing
-- servers will be skipped to avoid migration failure. If you need to keep
-- those rows, review and re-insert them after creating the corresponding
-- `servers` records.
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
  w.`context`, 
  w.`url` as `connection_string`,
  'webhook' as `type`,
  w.`forwardinternal`, 
  w.`trackid`, 
  w.`readreceipts`, 
  w.`groups`, 
  w.`broadcasts`, 
  w.`extra`, 
  w.`timestamp`
FROM `webhooks` w
INNER JOIN `servers` s ON s.`token` = w.`context`;

-- If you want to inspect webhooks that were NOT migrated because the
-- referenced `servers.token` doesn't exist, you can run the following
-- (manually, outside of this migration) to list them and decide how to
-- recover them. This is intentionally commented out to avoid execution
-- during automated migrations.
--
-- SELECT w.* FROM `webhooks` w
-- LEFT JOIN `servers` s ON s.`token` = w.`context`
-- WHERE s.`token` IS NULL;

-- Drop old webhooks table if exists
DROP TABLE IF EXISTS `webhooks`;

-- Drop old dispatching table if exists (in case of re-run)
DROP TABLE IF EXISTS `dispatching`;

-- Rename new table to dispatching
ALTER TABLE `dispatching_temp` RENAME TO `dispatching`;
