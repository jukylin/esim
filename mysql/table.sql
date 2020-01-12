create database test_1;
CREATE TABLE IF NOT EXISTS test_1.test(
  id int not NULL DEFAULT 0,
  title VARCHAR(10) not NULL DEFAULT ''
)engine=innodb;

create database test_2;
CREATE TABLE IF NOT EXISTS test_2.user(
  id int not NULL DEFAULT 0,
  username VARCHAR(10) not NULL DEFAULT ''
)engine=innodb;