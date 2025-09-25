package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 区域-场景
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_area_scene")
public class StardewAreaSceneEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 关联的区域ID
     */
    private Integer areaId;

    /**
     * 编码
     */
    private String sysCode;

    /**
     * 数据类型
     */
    private Integer dataMode;

    /**
     * 场景名称
     */
    private String nameCh;

    /**
     * 居民
     */
    private String people;

    /**
     * 开放时间
     */
    private String openTime;

    /**
     * 关闭时间
     */
    private String closeTime;

    /**
     * 地址
     */
    private String address;

    /**
     * 描述
     */
    private String remark;
}
