-- Add readupdate column to servers table
-- Using ALTER TABLE ADD COLUMN to avoid foreign key constraint issues
-- (dispatching table references servers.token)
ALTER TABLE `servers` ADD COLUMN `readupdate` INT(1) NOT NULL DEFAULT 0;
