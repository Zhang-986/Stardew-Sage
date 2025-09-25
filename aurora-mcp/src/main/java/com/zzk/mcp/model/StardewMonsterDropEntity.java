package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 怪物-掉落
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_monster_drop")
public class StardewMonsterDropEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 怪物ID
     */
    private String monsterId;

    /**
     * 物品名称
     */
    private String materialName;

    /**
     * 掉落概率
     */
    private String dropRate;

    /**
     * 是否为链接
     */
    private Boolean linkIs;
}
