package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import java.math.BigDecimal;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 鱼塘升级和产出
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_fish_product")
public class StardewFishProductEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 关联鱼类Code
     */
    private Integer fishId;

    /**
     * 需要数量
     */
    private Integer needCount;

    /**
     * 产出和升级材料
     */
    private String material;

    /**
     * 概率
     */
    private BigDecimal rate;

    /**
     * 01产出，02升级
     */
    private String recordType;
}
