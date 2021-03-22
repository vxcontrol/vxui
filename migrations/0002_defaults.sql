-- +migrate Up

SET SESSION sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

INSERT IGNORE INTO `groups` (`id`, `name`) VALUES
	(0, 'SAdmin'),
	(1, 'Admin'),
    (2, 'User');

INSERT IGNORE INTO `tenants` (`id`, `status`) VALUES
	(0, 'active');

INSERT IGNORE INTO `users` (`id`, `mail`, `name`, `password`, `status`, `group_id`, `tenant_id`) VALUES
	(0, 'vxadmin', 'VX Admin', '$2a$10$RjtRo/2.sxh.KUBSvryuB.MiNN94Tqrc7ikMEXFaUnCEFT0an3Nke', 'active', 0, 0);

-- +migrate Down

DELETE FROM `users` WHERE `id` IN (0);
DELETE FROM `tenants` WHERE `id` IN (0);
DELETE FROM `groups` WHERE `id` IN (0, 1, 2);
