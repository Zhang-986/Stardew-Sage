package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 畜产品
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_husbandry")
public class StardewHusbandryEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 编码
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
     * 普通价格
     */
    private String normalPrice;

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
     * 来源
     */
    private String sourceFrom;

    /**
     * 加工产品
     */
    private String product;

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
