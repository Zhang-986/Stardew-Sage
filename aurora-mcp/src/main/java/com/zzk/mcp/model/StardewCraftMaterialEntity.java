package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 工艺-材料
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_craft_material")
public class StardewCraftMaterialEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 工艺ID
     */
    private String craftId;

    /**
     * 材料名称
     */
    private String materialName;

    /**
     * 需要数量
     */
    private Integer needNum;

    /**
     * 是否为链接
     */
    private Boolean linkIs;
}
