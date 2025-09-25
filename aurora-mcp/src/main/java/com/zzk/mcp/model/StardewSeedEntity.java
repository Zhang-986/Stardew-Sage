package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import java.math.BigDecimal;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 种子
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_seed")
public class StardewSeedEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 种子编码
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
     * 生长季节
     */
    private String season;

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
     * 作物
     */
    private String cropName;

    /**
     * 作物售价
     */
    private String sellPrice;

    /**
     * 种子价格
     */
    private String buyPrice;

    /**
     * 最小收获
     */
    private Integer minHarvest;

    /**
     * 最大收获
     */
    private Integer maxHarvest;

    /**
     * 概率
     */
    private BigDecimal harvestRate;

    /**
     * 备注
     */
    private String remark;

    /**
     * 来源
     */
    private String sourceFrom;

    /**
     * 能否计算
     */
    private Integer calculatorIs;
}
