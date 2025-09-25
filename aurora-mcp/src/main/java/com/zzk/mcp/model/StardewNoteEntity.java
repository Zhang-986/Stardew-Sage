package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 描述信息
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_note")
public class StardewNoteEntity {

      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 关联类型
     */
    private String relType;

    /**
     * 编码
     */
    private String sysCode;

    /**
     * 标题
     */
    private String title;

    /**
     * 数据类型
     */
    private String type;

    /**
     * 描述内容
     */
    private String content;

    /**
     * 图片模式
     */
    private String imageMode;
}
