use Verification;
-- -----------------------------------
-- 更新或插入实名认证信息
-- -----------------------------------
SET SQL_SAFE_UPDATES = 0;
drop procedure if exists verificationUpdate;

Delimiter &&

create procedure verificationUpdate(
	in acc varchar(60),
    in lastLgTime varchar(60),	-- 00000000-月份00-天数00-小时00-分钟00
    in lastEtTime varchar(60),
	in id varchar(60),	-- 身份证号码
    in canGo int,	-- 1:可以继续游戏,2不可以继续游戏
    in ag int,   -- 用户是否认证为未成年	in isRelVerify int -- 用户是否已经认证
    in purchaseNum int -- 用户剩余的购买额度
)
begin
	declare num int;
	 declare t_error integer default 0; 
	 declare continue handler for sqlexception set t_error=1;
     start transaction;
	    select count(*) into num from verificationSystem where accoun=acc;
		if num=0 then
			insert into verificationSystem(accoun, lastLoginTime, lastExitTime, userId,canGoOn , age,residuePurcharse)values(acc, lastLgTime , lastEtTime , id , canGo , ag,purchaseNum );
		else
			update verificationSystem set lastLoginTime = lastLgTime , lastExitTime = lastEtTime, userId = id, canGoOn = canGo , isYoung = ag,residuePurcharse=purchaseNum where accoun = acc;
	   	end if;		
    if t_error = 1 then 
		rollback; 
	 else 
		commit; 
	end if;
   --  select * from verificationSystem where accoun=acc;   
end &&

Delimiter ;

-- -----------------------------------
-- 查询实名信息
-- -----------------------------------

drop  procedure if exists verifyRepeatIdty;

Delimiter &&

create procedure verifyRepeatIdty(
		in acc varchar(60)
)
begin 
	select * from verificationSystem where accoun=acc;
end &&

Delimiter ;

-- -----------------------------------
-- 查询信息是否存在
-- -----------------------------------

drop  procedure if exists verifyReallyHas;

Delimiter &&

create procedure verifyReallyHas(
		in cardId varchar(60)
)
begin 
	declare num int;
	select  count(*) into num from verificationSystem where accoun=cardId;
    select num;
end &&

Delimiter ;
-- -----------------------------------
-- 查询信息是否被认证
-- -----------------------------------

drop  procedure if exists verifyRepeat;

Delimiter &&

create procedure verifyRepeat(
		
        in gameSign varchar(60),	-- 游戏标示
		in cardId varchar(60)	-- 身份证号码
      
)
begin 
	declare num int;
	select  count(*) into num from verificationSystem where accoun like  CONCAT(gameSign,'%') and userId=cardId ;
    select num;
end &&

Delimiter ;

-- -----------------------------------
-- 每个月更新一次数据库中未成年的剩余消费额度
-- -----------------------------------
drop procedure if exists updatePurchaseNum;

Delimiter &&

create procedure updatePurchaseNum()
begin
	 declare _error int default 0;
	 declare continue handler  for sqlexception set  _error=1; 
     start transaction ;
    update verificationSystem set residuePurcharse=200 where age between 8 and 15 ;
    update verificationSystem set residuePurcharse=400 where age between 16 and 17;
    if _error=1 then
      rollback;
	else
     commit;
	end if;
end &&
Delimiter ;
