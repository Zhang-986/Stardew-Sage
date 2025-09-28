package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

import java.math.BigDecimal;

/**
 * <p>
 * 装备
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_equip")
public class StardewEquipEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 代码
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
     * 类型一
     */
    private String equipTypeOne;

    /**
     * 类型二
     */
    private String equipTypeTwo;

    /**
     * 等级
     */
    private Integer level;

    /**
     * 武器攻击力（最低）
     */
    private Integer attackDown;

    /**
     * 武器攻击力（最高）
     */
    private Integer attackUp;

    /**
     * 武器暴击率
     */
    private BigDecimal equipCrit;

    /**
     * 速度
     */
    private Integer speed;

    /**
     * 防御
     */
    private Integer def;

    /**
     * 重量
     */
    private Integer weight;

    /**
     * 幸运
     */
    private Integer luck;

    /**
     * 暴击率
     */
    private BigDecimal critRate;

    /**
     * 暴击力量
     */
    private Integer critPower;

    /**
     * 免疫
     */
    private Integer immunity;

    /**
     * 效果
     */
    private String effect;

    /**
     * 描述
     */
    private String remark;

    /**
     * 来源
     */
    private String sourceFrom;

    /**
     * 购买价格
     */
    private String priceBuy;

    /**
     * 出售价格
     */
    private String priceSell;
}
