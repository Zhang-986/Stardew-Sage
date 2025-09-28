package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

import java.math.BigDecimal;

/**
 * <p>
 * 鱼塘产出
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_fish_output")
public class StardewFishOutputEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 关联ID
     */
    private Integer fishId;

    /**
     * 产出物品
     */
    private String goodsName;

    /**
     * 数量
     */
    private String goodsNum;

    /**
     * 一条基础概率
     */
    private BigDecimal baseOne;

    /**
     * 二条基础概率
     */
    private BigDecimal baseTwo;

    /**
     * 三条基础概率
     */
    private BigDecimal baseThree;

    /**
     * 四条基础概率
     */
    private BigDecimal baseFour;

    /**
     * 五条基础概率
     */
    private BigDecimal baseFive;

    /**
     * 六条基础概率
     */
    private BigDecimal baseSix;

    /**
     * 七条基础概率
     */
    private BigDecimal baseSeven;

    /**
     * 八条基础概率
     */
    private BigDecimal baseEight;

    /**
     * 九条基础概率
     */
    private BigDecimal baseNine;

    /**
     * 十条基础概率
     */
    private BigDecimal baseTen;

    /**
     * 一条实际概率
     */
    private Integer realOne;

    /**
     * 二条实际概率
     */
    private Integer realTwo;

    /**
     * 三条实际概率
     */
    private Integer realThree;

    /**
     * 四条实际概率
     */
    private Integer realFour;

    /**
     * 五条实际概率
     */
    private Integer realFive;

    /**
     * 六条实际概率
     */
    private Integer realSix;

    /**
     * 七条实际概率
     */
    private Integer realSeven;

    /**
     * 八条实际概率
     */
    private Integer realEight;

    /**
     * 九条实际概率
     */
    private Integer realNine;

    /**
     * 十条实际概率
     */
    private Integer realTen;
}
