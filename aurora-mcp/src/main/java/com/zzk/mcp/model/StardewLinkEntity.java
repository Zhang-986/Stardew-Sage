package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 链接
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_link")
public class StardewLinkEntity {

      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 链接内容
     */
    private String linkText;

    /**
     * 链接类型
     */
    private Integer linkType;

    /**
     * 链接参数
     */
    private String linkValue;

    /**
     * 链接图标
     */
    private String linkIcon;

    /**
     * 数据类型
     */
    private Integer dataMode;

    /**
     * 是否为链接
     */
    private Boolean linkIs;
}
