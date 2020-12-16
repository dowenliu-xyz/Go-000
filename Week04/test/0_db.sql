CREATE DATABASE crm;
USE crm;
CREATE TABLE IF NOT EXISTS customer_follow
(
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `staff_id`    BIGINT          NOT NULL COMMENT '跟进客户的员工id，记录日志的员工id',
    `customer_id` BIGINT          NOT NULL COMMENT '被跟进客户的id',
    `content`     TEXT            NOT NULL COMMENT '跟进日志内容',
    `create_time` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '跟进日志创建日期',
    `deleted`     BIGINT          NOT NULL DEFAULT 0 COMMENT '跟进日志删除时间戳。如果为0表示未删除',
    `deleted_by`  VARCHAR(50)     NULL     DEFAULT NULL COMMENT '删除人。可能是员工自己删除，也可能是系统删除、管理员删除',
    PRIMARY KEY (id),
    INDEX (deleted, staff_id, create_time DESC),
    INDEX (deleted, customer_id, staff_id, create_time DESC)
) COMMENT = '客户跟进日志表';


