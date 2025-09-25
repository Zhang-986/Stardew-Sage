package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 收集包内容
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_bundle_content")
public class StardewBundleContentEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 关联收集包CODE
     */
    private String sysCode;

    /**
     * 收集包内容
     */
    private String contentName;

    /**
     * 收集包描述
     */
    private String contentNote;

    /**
     * 是否混合模式独有
     */
    private Boolean contentStatus;
}
