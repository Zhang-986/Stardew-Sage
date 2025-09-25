package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 村民
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_people")
public class StardewPeopleEntity {

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
     * 生日
     */
    private String birthday;

    /**
     * 住址
     */
    private String address;

    /**
     * 结婚对象
     */
    private String married;

    /**
     * 性别
     */
    private String personSex;

    /**
     * 家庭成员
     */
    private String familyMembers;

    /**
     * 其他信息
     */
    private String otherInfo;

    /**
     * 首字母
     */
    private String groupWord;

    /**
     * 最爱的电影
     */
    private String favoriteMovie;

    /**
     * 最爱的零食
     */
    private String favoriteNibble;

    /**
     * 数据模式
     */
    private Integer dataMode;

    /**
     * 特殊标志
     */
    private String spTag;

    /**
     * 备注
     */
    private String remark;
}
