package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 矿物
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_mineral")
public class StardewMineralEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 矿物名称-中文
     */
    private String nameCh;

    /**
     * 矿物编码
     */
    private String sysCode;

    /**
     * 价格
     */
    private String sellPrice;

    /**
     * 描述
     */
    private String remark;

    /**
     * 用途
     */
    private String purpose;

    /**
     * 获取途径
     */
    private String sourceFrom;

    /**
     * 数据模式
     */
    private Integer dataMode;

    /**
     * 矿物类型
     */
    private String mineralType;
}
