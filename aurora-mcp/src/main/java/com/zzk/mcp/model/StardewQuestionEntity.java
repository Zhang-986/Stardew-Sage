package com.zzk.mcp.model;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 题目
 * </p>
 *
 * @author ZZK
 * @since 2025-09-25
 */
@Getter
@Setter
@TableName("stardew_question")
public class StardewQuestionEntity {

    /**
     * 主键
     */
      @TableId(value = "id", type = IdType.AUTO)
    private Integer id;

    /**
     * 数据类型
     */
    private Integer dataMode;

    /**
     * 题目内容
     */
    private String questionContent;

    /**
     * 题目选项A
     */
    private String optionA;

    /**
     * 题目选项B
     */
    private String optionB;

    /**
     * 题目选项C
     */
    private String optionC;

    /**
     * 题目选项D
     */
    private String optionD;

    /**
     * 答案
     */
    private String questionAnswer;

    /**
     * 解析
     */
    private String questionAnalysis;
}
