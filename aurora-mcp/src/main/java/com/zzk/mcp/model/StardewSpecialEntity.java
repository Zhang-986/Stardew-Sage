package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 特殊物品
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_special")
public class StardewSpecialEntity {

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
     * 名称-中文
     */
    private String nameCh;

    /**
     * 数据模式
     */
    private Integer dataMode;

    /**
     * 售卖价格
     */
    private String sellPrice;

    /**
     * 购买价格
     */
    private String buyPrice;

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
