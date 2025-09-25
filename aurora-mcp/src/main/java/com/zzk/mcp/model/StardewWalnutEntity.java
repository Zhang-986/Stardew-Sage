package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 金色核桃
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_walnut")
public class StardewWalnutEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 名称-中文
     */
    private String nameCh;

    /**
     * 描述
     */
    private String remark;

    /**
     * 核桃数量
     */
    private Integer walnutNum;

    /**
     * 类型
     */
    private String type;
}
