package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 料理-材料
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_food_material")
public class StardewFoodMaterialEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 料理ID
     */
    private String foodId;

    /**
     * 材料名称
     */
    private String materialName;

    /**
     * 材料编码
     */
    private String materialCode;

    /**
     * 需要数量
     */
    private Integer needNum;

    /**
     * 是否为链接
     */
    private Boolean linkIs;
}
