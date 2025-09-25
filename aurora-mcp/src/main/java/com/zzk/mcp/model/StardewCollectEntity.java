package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 全可收集物品
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_collect")
public class StardewCollectEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 关联编码
     */
    private String relationCode;

    /**
     * 名称
     */
    private String nameCh;

    /**
     * 数据模式
     */
    private Integer dataMode;

    /**
     * 所属分类
     */
    private String type;
}
