-- +migrate Up

CREATE TABLE IF NOT EXISTS `tenants` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `status` ENUM('active','blocked') NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `modules` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` INT(10) UNSIGNED NOT NULL,
  `service_type` ENUM('vxmonitor') NOT NULL,
  `config_schema` JSON NOT NULL,
  `default_config` JSON NOT NULL,
  `event_data_schema` JSON NOT NULL,
  `event_config_schema` JSON NOT NULL,
  `default_event_config` JSON NOT NULL,
  `changelog` JSON NOT NULL,
  `locale` JSON NOT NULL,
  `info` JSON NOT NULL,
  `name` VARCHAR(50) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.name'))) STORED,
  `version` VARCHAR(10) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.version'))) STORED,
  `tags` JSON GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.tags'))) VIRTUAL,
  `events` JSON GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.events'))) VIRTUAL,
  `template` VARCHAR(50) GENERATED ALWAYS AS (json_extract(`info`,_utf8mb4'$.template')) STORED,
  `system` TINYINT(1) GENERATED ALWAYS AS (json_extract(`info`,_utf8mb4'$.system')) STORED,
  `last_update` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `fkm_tenant_id` (`tenant_id`),
  CONSTRAINT `fkm_tenant_id` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `groups` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `services` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` INT(10) UNSIGNED NOT NULL,
  `name` VARCHAR(50) NOT NULL,
  `type` ENUM('vxmonitor') NOT NULL,
  `status` ENUM('created','active','blocked','removed') NOT NULL,
  `info` JSON NOT NULL,
  `db_name` VARCHAR(50) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.db.name'))) STORED,
  `db_user` VARCHAR(50) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.db.user'))) STORED,
  `db_pass` VARCHAR(50) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.db.pass'))) STORED,
  `db_host` VARCHAR(50) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.db.host'))) STORED,
  `db_port` INT(10) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.db.port'))) STORED,
  `server_host` VARCHAR(50) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.server.host'))) STORED,
  `server_port` INT(10) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.server.port'))) STORED,
  `server_proto` VARCHAR(10) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,_utf8mb4'$.server.proto'))) STORED,
  `s3_endpoint` VARCHAR(100) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,'$.s3.endpoint'))) STORED,
  `s3_access_key` VARCHAR(50) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,'$.s3.access_key'))) STORED,
  `s3_secret_key` VARCHAR(50) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,'$.s3.secret_key'))) STORED,
  `s3_bucket_name` VARCHAR(30) GENERATED ALWAYS AS (json_unquote(json_extract(`info`,'$.s3.bucket_name'))) STORED,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`tenant_id`, `name`, `type`),
  KEY `fks_tenant_id` (`tenant_id`),
  CONSTRAINT `fks_tenant_id` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `users` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `mail` VARCHAR(50) NOT NULL,
  `name` VARCHAR(70) NOT NULL DEFAULT '',
  `password` VARCHAR(100) NOT NULL,
  `status` ENUM('created','active','blocked') NOT NULL,
  `group_id` INT(10) UNSIGNED NOT NULL DEFAULT '2',
  `tenant_id` INT(10) UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `mail` (`mail`),
  KEY `fku_group_id` (`group_id`),
  KEY `fku_tenant_id` (`tenant_id`),
  CONSTRAINT `fku_group_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`),
  CONSTRAINT `fku_tenant_id` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8;

-- +migrate Down

SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `groups`;
DROP TABLE IF EXISTS `services`;
DROP TABLE IF EXISTS `modules`;
DROP TABLE IF EXISTS `tenants`;
