package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 村民-关系
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_people_relation")
public class StardewPeopleRelationEntity {

    /**
     * 主键
     */
    private Integer id;

    /**
     * 村民ID
     */
    private Integer personOneId;
}
