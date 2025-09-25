package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 工匠物品
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_product")
public class StardewProductEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 工匠物品编码
     */
    private String sysCode;

    /**
     * 名称-中文
     */
    private String nameCh;

    /**
     * 原料
     */
    private String material;

    /**
     * 普通价格
     */
    private String normalPrice;

    /**
     * 售出价格
     */
    private String sellPrice;

    /**
     * 生产耗时
     */
    private String productionTime;

    /**
     * 描述
     */
    private String remark;

    /**
     * 配方来源
     */
    private String productFrom;

    /**
     * 设备
     */
    private String device;

    /**
     * 能量
     */
    private String energy;

    /**
     * 生命
     */
    private String health;

    /**
     * 数据模式
     */
    private Integer dataMode;
}
