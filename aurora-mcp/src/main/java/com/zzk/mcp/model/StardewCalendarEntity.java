package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 日历
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_calendar")
public class StardewCalendarEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 季节
     */
    private String season;

    /**
     * 节日
     */
    private String festival;

    /**
     * 日期
     */
    private Integer dayNum;

    /**
     * 节日类型
     */
    private String festivalType;

    /**
     * 关联CODE
     */
    private String relationCode;

    /**
     * 数据模式
     */
    private Integer dataMode;
}
