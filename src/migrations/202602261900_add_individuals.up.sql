-- Add individuals option to servers and dispatching
ALTER TABLE `servers` ADD COLUMN `individuals` INT(1) NOT NULL DEFAULT 0;
ALTER TABLE `dispatching` ADD COLUMN `individuals` INT(1) NOT NULL DEFAULT 0;
