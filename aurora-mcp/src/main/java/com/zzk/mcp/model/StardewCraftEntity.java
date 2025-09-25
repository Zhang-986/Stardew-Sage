package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 工艺
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_craft")
public class StardewCraftEntity {

    /**
     * 主键
     */
    @TableId(value = "id", type = IdType.NONE)
    private Integer id;

    /**
     * 数据类型
     */
    private Integer dataMode;

    /**
     * 编码
     */
    private String sysCode;

    /**
     * 名称-中文
     */
    private String nameCh;

    /**
     * 类型
     */
    private String craftType;

    /**
     * 打造价格
     */
    private String makePrice;

    /**
     * 售出价格
     */
    private String sellPrice;

    /**
     * 来源
     */
    private String sourceFrom;

    /**
     * 简介
     */
    private String remark;
}
