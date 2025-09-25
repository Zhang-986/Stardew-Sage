package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 采集品
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_harvest")
public class StardewHarvestEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 采集品编码
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
     * 季节
     */
    private String season;

    /**
     * 采集地点
     */
    private String place;

    /**
     * 普通价格
     */
    private Integer normalPrice;

    /**
     * 售卖价格
     */
    private String sellPrice;

    /**
     * 能量
     */
    private String energy;

    /**
     * 生命
     */
    private String health;

    /**
     * 是否能种植
     */
    private String harvestCrop;

    /**
     * 可食用性
     */
    private String edibility;

    /**
     * 种类
     */
    private String category;

    /**
     * 增益效果
     */
    private String buffStatus;

    /**
     * 增益时长
     */
    private String buffTime;

    /**
     * 用途
     */
    private String purpose;

    /**
     * 合成
     */
    private String composed;

    /**
     * 描述
     */
    private String remark;
}
