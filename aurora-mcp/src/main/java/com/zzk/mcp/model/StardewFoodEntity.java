package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 料理
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_food")
public class StardewFoodEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 料理代码
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
     * 能量
     */
    private Integer energy;

    /**
     * 生命
     */
    private Integer healthy;

    /**
     * 增益效果
     */
    private String buffStatus;

    /**
     * 增益时长
     */
    private String buffTime;

    /**
     * 食谱来源
     */
    private String sourceFrom;

    /**
     * 售出价格
     */
    private Integer sellPrice;

    /**
     * 描述
     */
    private String remark;

    /**
     * 用途
     */
    private String purpose;
}
