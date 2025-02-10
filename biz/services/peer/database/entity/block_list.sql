CREATE TABLE `block_list` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `target` varchar(40) NOT NULL COMMENT 'banned target',
  `type` tinyint(4) NOT NULL COMMENT 'banned type, 0 - info_hash, 1 - peer_id',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_target_type` (`type`,`target`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Black list for info_hash and peers'