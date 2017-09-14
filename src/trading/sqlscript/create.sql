create database if not exists trading;
use trading;
create table if not exists `trading_user` (
    `id` bigint(20) not null AUTO_INCREMENT,
    `money` int NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 auto_increment=10000 ;


create table if not exists `serial_number` (
    `id` bigint(20) not null AUTO_INCREMENT,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 auto_increment=100000 ;

create table if not exists `active_trading_data` (
  `serial_number` bigint(20) NOT NULL,
	`source_id` bigint(20) NOT NULL,
	`target_id` bigint(20) NOT null,
  `amount` int  NOT NULL ,
  `oper_type` int NOT NULL,
  `state` int NOT NULL,
	PRIMARY KEY (`serial_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


create table if not exists `passive_trading_data` (
  `serial_number` varchar(128),
	`source_id` bigint(20) NOT NULL,
	`target_id` bigint(20) NOT null,
  `amount` int  NOT NULL ,
  `oper_type` int NOT NULL,
  `state` int NOT NULL,
	PRIMARY KEY (`serial_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
