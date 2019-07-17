CREATE TABLE `identities_bsc` (
  `ip` varchar(45) CHARACTER SET utf8 COLLATE utf8_spanish_ci DEFAULT NULL,
  `provider` varchar(255) COLLATE utf8_spanish_ci DEFAULT NULL,
  `application` varchar(255) COLLATE utf8_spanish_ci DEFAULT NULL,
  `passport_id_grp` varchar(45) COLLATE utf8_spanish_ci DEFAULT NULL,
  `passport_id` varchar(45) COLLATE utf8_spanish_ci DEFAULT NULL,
  `createdat` datetime DEFAULT NULL,
  `ididentity` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`ididentity`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8 COLLATE=utf8_spanish_ci