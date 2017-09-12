create database if not exists trading;
use trading;
create table if not exists `trading_user` (
    `id` bigint(20) not null AUTO_INCREMENT,
    `money` int NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

create table if not exists `trading_data` (
  `id` bigint(20) not null AUTO_INCREMENT,
	`source_id` bigint(20) NOT NULL,
	`dest_id` bigint(20) NOT null,
  `amount` int  NOT NULL ,
  `oper_type` int NOT NULL,
	PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
