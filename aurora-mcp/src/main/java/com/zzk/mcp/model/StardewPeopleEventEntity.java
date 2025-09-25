package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 村民-事件
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_people_event")
public class StardewPeopleEventEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 村民ID
     */
    private Integer personId;

    /**
     * 事件名称
     */
    private String eventName;

    /**
     * 概要
     */
    private String noteSimple;

    /**
     * 详情
     */
    private String noteDetail;

    /**
     * 事件等级
     */
    private Integer eventLevel;

    /**
     * 时间
     */
    private String eventTime;

    /**
     * 地点
     */
    private String eventPlace;

    /**
     * 天气
     */
    private String weather;

    /**
     * 季节
     */
    private String season;

    /**
     * 额外条件
     */
    private String eventExtra;

    /**
     * 数据模式
     */
    private Integer dataMode;
}
