Create Database  IF NOT EXISTS `Verification` DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;

use Verification;
-- ----------------------------
-- Table structure for Verification`
-- ----------------------------
DROP TABLE IF EXISTS `verificationSystem`;
CREATE TABLE `verificationSystem` (	
	`accoun`                    varchar(60) NOT NULL DEFAULT "",					#用户ID
	`lastLoginTime`    			varchar(60) NOT NULL DEFAULT "",					#用户上次登陆的时间
	`lastExitTime`    			varchar(60) NOT NULL DEFAULT "",					#用户上次退出的时间
	`userId`					varchar(60) NOT NULL DEFAULT "",					#用户身份证号码
	`canGoOn`					int NOT NULL DEFAULT 0,								#用户是否能继续玩游戏
	`age`					    int NOT NULL DEFAULT 0,								#用户年龄
    `residuePurcharse`			int NOT NULL DEFAULT 0,								#剩余购买额度
	/*`isReallyVerify`			int NOT NULL DEFAULT 0,								#用户是否已经认证*/
	PRIMARY KEY (`accoun`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;