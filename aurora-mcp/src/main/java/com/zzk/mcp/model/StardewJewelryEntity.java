package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 饰品
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_jewelry")
public class StardewJewelryEntity {

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
     * 来源
     */
    private String sourceFrom;

    /**
     * 锻造条目
     */
    private String forgeEntry;

    /**
     * 锻造效果
     */
    private String forgeEffect;

    /**
     * 备注
     */
    private String remark;

    /**
     * 售出价格
     */
    private String sellPrice;

    /**
     * 购买价格
     */
    private String buyPrice;
}
