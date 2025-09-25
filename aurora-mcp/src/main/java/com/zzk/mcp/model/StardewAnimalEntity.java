package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 动物
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_animal")
public class StardewAnimalEntity {

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
     * 购买价格
     */
    private String buyPrice;

    /**
     * 售出价格
     */
    private String sellPrice;

    /**
     * 需要设施
     */
    private String needBuild;

    /**
     * 产出
     */
    private String animalOutput;

    /**
     * 备注
     */
    private String remark;

    /**
     * 获取方式
     */
    private String obtain;
}
