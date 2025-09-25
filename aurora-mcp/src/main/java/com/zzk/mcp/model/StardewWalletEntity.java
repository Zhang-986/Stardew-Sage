package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 钱包
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_wallet")
public class StardewWalletEntity {

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
     * 名称-中文
     */
    private String nameCh;

    /**
     * 类型
     */
    private String type;

    /**
     * 效果
     */
    private String propUsage;

    /**
     * 获取
     */
    private String obtain;

    private Integer dataMode;
}
