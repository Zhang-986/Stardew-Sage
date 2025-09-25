package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 树
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_tree")
public class StardewTreeEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 树编码
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
     * 类型
     */
    private String treeType;

    /**
     * 季节
     */
    private String season;

    /**
     * 生长周期
     */
    private String growthDay;

    /**
     * 用途
     */
    private String purpose;

    /**
     * 描述
     */
    private String remark;

    /**
     * 普通价格
     */
    private Integer normalPrice;

    /**
     * 售卖价格
     */
    private String sellPrice;

    /**
     * 来源
     */
    private String sourceFrom;

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
     * 树名称
     */
    private String treeName;

    /**
     * 树苗名称
     */
    private String seedName;

    /**
     * 水果名称
     */
    private String fruitName;
}
