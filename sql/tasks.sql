CREATE TABLE `tasks` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `customer_id` int(11) DEFAULT NULL,
  `executor_id` int(11) DEFAULT NULL,
  `title` varchar(45) DEFAULT NULL,
  `status` tinyint(4) NOT NULL DEFAULT '0',
  `cost` decimal(13,12) NOT NULL DEFAULT '0.000000000000',
  `problem` blob,
  `solution` blob,
  `begin_time` datetime DEFAULT '0001-01-01 00:00:00',
  `end_time` datetime DEFAULT '0001-01-01 00:00:00',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
