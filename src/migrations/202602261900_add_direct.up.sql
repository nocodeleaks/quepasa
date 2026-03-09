-- Add direct option to servers and dispatching
ALTER TABLE `servers` ADD COLUMN `direct` INT(1) NOT NULL DEFAULT 0;
ALTER TABLE `dispatching` ADD COLUMN `direct` INT(1) NOT NULL DEFAULT 0;
