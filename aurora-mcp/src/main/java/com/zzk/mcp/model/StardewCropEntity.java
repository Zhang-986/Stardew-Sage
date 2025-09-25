package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import java.math.BigDecimal;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 农作物
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_crop")
public class StardewCropEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 农作物编码
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
     * 普通价格
     */
    private Integer normalPrice;

    /**
     * 售卖价格
     */
    private String sellPrice;

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
     * 生长周期
     */
    private String growthDay;

    /**
     * 是否循环生长
     */
    private String cyclicIs;

    /**
     * 循环周期
     */
    private String cyclicDay;

    /**
     * 种子名称
     */
    private String seedName;

    /**
     * 作物种类
     */
    private String category;

    /**
     * 最小收获
     */
    private Integer minHarvest;

    /**
     * 最大收获
     */
    private Integer maxHarvest;

    /**
     * 收获概率
     */
    private BigDecimal harvestRate;

    /**
     * 收获经验
     */
    private Integer harvestExp;

    /**
     * 是否能采集
     */
    private String harvestCrop;

    /**
     * 是否镰刀收获
     */
    private String scytheIs;

    /**
     * 是否有架子
     */
    private String trellisIs;

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
