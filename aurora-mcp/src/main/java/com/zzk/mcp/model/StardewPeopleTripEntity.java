package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 村民-行程
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_people_trip")
public class StardewPeopleTripEntity {

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
     * 行程类型
     */
    private String tripType;

    /**
     * 类型顺序
     */
    private Integer sortOrderTrip;

    /**
     * 行程日期
     */
    private String dayTag;

    /**
     * 日期顺序
     */
    private Integer sortOrderDay;

    /**
     * 时间
     */
    private String timeTag;

    /**
     * 行为
     */
    private String behaviorTag;

    /**
     * 模组
     */
    private Integer dataMode;
}
