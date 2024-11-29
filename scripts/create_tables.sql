DROP TABLE IF EXISTS `box`;
CREATE TABLE `box` (
`username`  varchar(255) NOT NULL DEFAULT '',
`boxId` varchar(255) NOT NULL DEFAULT '',
`supplierBoxId` varchar(255) NOT NULL DEFAULT '',
`online` varchar(255) NOT NULL DEFAULT '',
`tcpNatType` varchar(255) NOT NULL DEFAULT '',
`udpNatType` varchar(255) NOT NULL DEFAULT '',
`publicIp` varchar(255) NOT NULL DEFAULT '',
`privateIp` varchar(255) NOT NULL DEFAULT '',
`isp` varchar(255) NOT NULL DEFAULT '',
`province` varchar(255) NOT NULL DEFAULT '',
`city` varchar(255) NOT NULL DEFAULT '',
`cpuArch` varchar(255) NOT NULL DEFAULT '',
`cpuCores` varchar(255) NOT NULL DEFAULT '',
`memorySize` varchar(255) NOT NULL DEFAULT '',
`os` varchar(255) NOT NULL DEFAULT '',
`pluginVersion` varchar(255) NOT NULL DEFAULT '',
`pluginDeployTime` varchar(255) NOT NULL DEFAULT '',
`processStatus` varchar(255) NOT NULL DEFAULT '',
`fault` varchar(255) NOT NULL DEFAULT '',
`upload` float(32) NOT NULL DEFAULT 0,
`download` float(32) NOT NULL DEFAULT 0,
`diskUsage` float(32) NOT NULL DEFAULT 0,
`upnp` tinyint NOT NULL DEFAULT 0,
`notDeployReason` varchar(255) NOT NULL DEFAULT '',
`reportUpBandwidth` varchar(255) NOT NULL DEFAULT '',
`planTask` varchar(255) NOT NULL DEFAULT '',
`pressBandwidth` varchar(255) NOT NULL DEFAULT '',
`remark` varchar(255) NOT NULL DEFAULT '',
`icmpv6Out` float(32) NOT NULL DEFAULT 0,
`createdAt` datetime(3) NOT NULL DEFAULT 0,
`updatedAt` datetime(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`boxId`)
);

DROP TABLE IF EXISTS `box_diskinfo`;
CREATE TABLE `box_diskinfo` (
`boxId` varchar(255) NOT NULL DEFAULT '',
`supplierBoxId` varchar(255) NOT NULL DEFAULT '',
`diskId` varchar(255) NOT NULL DEFAULT '',
`diskSize` varchar(255) NOT NULL DEFAULT '',
`diskMedia` varchar(255) NOT NULL DEFAULT '',
`diskUsed` varchar(255) NOT NULL DEFAULT '',
INDEX `idx_boxId` USING BTREE(`boxId`),
UNIQUE KEY `uniq_boxid_diskid` (`boxId`, `diskId`) USING BTREE
);

DROP TABLE IF EXISTS `box_income`;
CREATE TABLE `box_income` (
username  varchar(255) NOT NULL DEFAULT '',
boxId varchar(255) NOT NULL DEFAULT '',
supplierBoxId varchar(255) NOT NULL DEFAULT '',
date varchar(255) NOT NULL DEFAULT '',
remark varchar(255) NOT NULL DEFAULT '',
bw varchar(255) NOT NULL DEFAULT '',
bwAmount varchar(255) NOT NULL DEFAULT '',
amount varchar(255) NOT NULL DEFAULT '',
activityIncome varchar(255) NOT NULL DEFAULT '',
distPercent varchar(255) NOT NULL DEFAULT '',
inviterId varchar(255) NOT NULL DEFAULT '',
updatedAt datetime(3) NOT NULL DEFAULT 0,
INDEX `idx_boxId_date` USING BTREE(`boxId`, `date`),
INDEX `idx_remark` USING BTREE(`remark`),
UNIQUE KEY `uniq_boxid_date` (`boxId`, `date`) USING BTREE
);

DROP TABLE IF EXISTS `box_bandwidth`;
CREATE TABLE `box_bandwidth` (
username  varchar(255) NOT NULL DEFAULT '',
boxId varchar(255) NOT NULL DEFAULT '',
supplierBoxId varchar(255) NOT NULL DEFAULT '',
time varchar(255) NOT NULL DEFAULT '',
upload float(32) NOT NULL DEFAULT 0,
download float(32) NOT NULL DEFAULT 0,
updatedAt datetime(3) NOT NULL DEFAULT 0,
INDEX `idx_boxId_time` USING BTREE(`boxId`, `time`),
UNIQUE KEY `uniq_boxid_time` (`boxId`, `time`) USING BTREE
);

DROP TABLE IF EXISTS `box_quality`;
CREATE TABLE `box_quality` (
username  varchar(255) NOT NULL DEFAULT '',
boxId varchar(255) NOT NULL DEFAULT '',
supplierBoxId varchar(255) NOT NULL DEFAULT '',
time varchar(255) NOT NULL DEFAULT '',
packetLoss varchar(255) NOT NULL DEFAULT '',
tcpNatType varchar(255) NOT NULL DEFAULT '',
udpNatType varchar(255) NOT NULL DEFAULT '',
cpuUsage float(32) NOT NULL DEFAULT 0,
memoryUsage float(32) NOT NULL DEFAULT 0,
diskUsage float(32) NOT NULL DEFAULT 0,
updatedAt datetime(3) NOT NULL DEFAULT 0,
INDEX `idx_boxId_time` USING BTREE(`boxId`, `time`),
UNIQUE KEY `uniq_boxid_time` (`boxId`, `time`) USING BTREE
);


DROP TABLE IF EXISTS `pai_userkey`;
CREATE TABLE `pai_userkey` (
paiUsername varchar(255) NOT NULL DEFAULT '',
username varchar(255) NOT NULL DEFAULT '',
apiKey varchar(255) NOT NULL DEFAULT '',
apiSecret varchar(255) NOT NULL DEFAULT '',
status tinyint(4) not null default 0
);

DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
uid BIGINT(20) NOT NULL AUTO_INCREMENT,
username varchar(255) NOT NULL DEFAULT '',
password varchar(255) NOT NULL DEFAULT '',
appKey varchar(255) NOT NULL DEFAULT '',
appSecret varchar(255) NOT NULL DEFAULT '',
supplierType bigint(20) not null default 0,
phoneNumber varchar(255) NOT NULL DEFAULT '',
billingCycle varchar(255) NOT NULL DEFAULT '',
parentId varchar(255) NOT NULL DEFAULT '',
distPercent bigint(20) not null default 0,
canInvite tinyint(4) not null default 0,
inviterType tinyint(4) not null default 0,
createdAt datetime(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`uid`)
)ENGINE=InnoDB AUTO_INCREMENT=100000;

