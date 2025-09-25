package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 古物
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_antique")
public class StardewAntiqueEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 古物编码
     */
    private String sysCode;

    /**
     * 数据模式
     */
    private Integer dataMode;

    /**
     * 名称-中文
     */
    private String nameCh;

    /**
     * 价格
     */
    private String sellPrice;

    /**
     * 描述
     */
    private String remark;

    /**
     * 获取途径
     */
    private String sourceFrom;

    /**
     * 远古斑点
     */
    private String spots;

    /**
     * 捐赠奖励
     */
    private String reward;
}
