ALTER TABLE `servers` ADD COLUMN `contextid` TEXT DEFAULT NULL;
CREATE INDEX IF NOT EXISTS `idx_servers_contextid` ON `servers` (`contextid`);
