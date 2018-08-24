CREATE TABLE `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `is_admin` tinyint(4) NOT NULL DEFAULT '0',
  `user_name` varchar(45) NOT NULL,
  `password_hash` varchar(45) NOT NULL,
  `email` varchar(45) DEFAULT NULL,
  `balance` decimal(13,12) NOT NULL DEFAULT '0.000000000000',
  `frozen_amount` decimal(13,12) NOT NULL DEFAULT '0.000000000000',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  UNIQUE KEY `user_name_UNIQUE` (`user_name`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
