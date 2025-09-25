package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 村民-爱好
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_people_favorite")
public class StardewPeopleFavoriteEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 村民ID
     */
    private Integer personId;

    /**
     * 爱好类型
     */
    private String favType;

    /**
     * 爱好名称
     */
    private String favName;

    /**
     * 数据模式
     */
    private Integer dataMode;

    /**
     * 01添加，02删除
     */
    private String favStatus;
}
