package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 秘密纸条
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_paper")
public class StardewPaperEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 编码
     */
    private String sysCode;

    /**
     * 数据类型
     */
    private Integer dataMode;

    /**
     * 纸条名称
     */
    private String nameCh;

    /**
     * 图片URL
     */
    private String imageUrl;

    /**
     * 纸条内容
     */
    private String paperNote;

    /**
     * 提示
     */
    private String paperTip;

    /**
     * 额外秘密
     */
    private String paperOther;

    /**
     * 备注
     */
    private String remark;
}
