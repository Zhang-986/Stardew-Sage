package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 收集包
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_bundle")
public class StardewBundleEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 代码
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
     * 解锁奖励
     */
    private String unlockReward;

    /**
     * 奖励
     */
    private String reward;

    /**
     * 奖励图标
     */
    private String rewardIcon;

    /**
     * 需要数量
     */
    private Integer needCount;

    /**
     * 备注
     */
    private String remark;

    /**
     * 收集包类型 01固定，02混合
     */
    private String bundleType;
}
