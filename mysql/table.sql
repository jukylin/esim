create database test_1;
CREATE TABLE IF NOT EXISTS test_1.test(
  id int not NULL auto_increment,
  title VARCHAR(10) not NULL DEFAULT '',
  PRIMARY KEY (`id`)
)engine=innodb;

create database test_2;
CREATE TABLE IF NOT EXISTS test_2.user(
  id int not NULL auto_increment,
  username VARCHAR(10) not NULL DEFAULT '',
    PRIMARY KEY (`id`)
)engine=innodb;