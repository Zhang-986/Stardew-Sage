package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 鱼类
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_fish")
public class StardewFishEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 鱼类代码
     */
    private String sysCode;

    /**
     * 数据模式
     */
    private Integer dataMode;

    /**
     * 名称-中文
     */
    private String nameCh;

    /**
     * 用途
     */
    private String purpose;

    /**
     * 出现季节
     */
    private String season;

    /**
     * 出现地点
     */
    private String appearPlace;

    /**
     * 出现时间
     */
    private String appearTime;

    /**
     * 天气
     */
    private String weather;

    /**
     * 难度
     */
    private Integer diffLevel;

    /**
     * 行为
     */
    private String behavior;

    /**
     * 普通价格
     */
    private Integer normalPrice;

    /**
     * 售卖价格
     */
    private String sellPrice;

    /**
     * 描述
     */
    private String remark;

    /**
     * 合成
     */
    private String composed;

    /**
     * 传说之鱼
     */
    private Boolean legend;

    /**
     * 分类
     */
    private String category;

    /**
     * 可食用性
     */
    private String edibility;

    /**
     * 能量
     */
    private String energy;

    /**
     * 生命
     */
    private String health;

    /**
     * 增益效果
     */
    private String buffStatus;

    /**
     * 增益时长
     */
    private String buffTime;
}
