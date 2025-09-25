package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 任务
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_mission")
public class StardewMissionEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 任务编码
     */
    private String sysCode;

    /**
     * 数据模式
     */
    private Integer dataMode;

    /**
     * 任务名称
     */
    private String nameCh;

    /**
     * 任务类型
     */
    private String missionType;

    /**
     * 任务内容
     */
    private String content;

    /**
     * 前置条件
     */
    private String missionCondition;

    /**
     * 任务需要
     */
    private String need;

    /**
     * 任务奖励
     */
    private String award;

    /**
     * 时间限制
     */
    private String timeOut;

    /**
     * 是否能重复
     */
    private Integer repeatIs;

    /**
     * 描述
     */
    private String remark;

    /**
     * 委托人
     */
    private String consignor;
}
