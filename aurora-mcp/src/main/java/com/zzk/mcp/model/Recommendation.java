package com.zzk.mcp.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.Map;

/**
 * 推荐结果实体
 * 
 * @author Stardew Sage Team
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class Recommendation {
    
    /**
     * 推荐项ID
     */
    private String itemId;
    
    /**
     * 推荐项名称
     */
    private String itemName;
    
    /**
     * 推荐类型（crop, character, recipe, etc.）
     */
    private String itemType;
    
    /**
     * 推荐分数（0-1）
     */
    private Double score;
    
    /**
     * 推荐原因
     */
    private String reason;
    
    /**
     * 额外的元数据
     */
    private Map<String, Object> metadata;
    
    /**
     * 推荐来源（collaborative, content_based, hybrid, ai_generated）
     */
    private String source;
}
