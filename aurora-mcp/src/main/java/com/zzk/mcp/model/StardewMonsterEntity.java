package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 怪物
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_monster")
public class StardewMonsterEntity {

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
     * 怪物代码
     */
    private String sysCode;

    /**
     * 生成地点
     */
    private String genPlace;

    /**
     * 生成地点编码
     */
    private String genPlaceCode;

    /**
     * 生成地点详情
     */
    private String genPlaceDetail;

    /**
     * 生命
     */
    private String hp;

    /**
     * 伤害
     */
    private String hurt;

    /**
     * 防御
     */
    private String def;

    /**
     * 速度
     */
    private String speed;

    /**
     * 经验
     */
    private String experience;

    /**
     * 描述
     */
    private String remark;

    /**
     * 数据模式
     */
    private Integer dataMode;

    /**
     * 可否杀死
     */
    private String killable;
}
