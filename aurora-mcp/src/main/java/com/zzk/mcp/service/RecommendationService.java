package com.zzk.mcp.service;

import com.zzk.mcp.model.Recommendation;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.document.Document;
import org.springframework.ai.vectorstore.SearchRequest;
import org.springframework.ai.vectorstore.VectorStore;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Flux;

import java.util.*;
import java.util.stream.Collectors;

/**
 * 智能推荐服务
 * 基于向量相似度和 AI 分析的混合推荐系统
 * 
 * @author Stardew Sage Team
 */
@Slf4j
@Service
public class RecommendationService {
    
    @Autowired(required = false)
    private VectorStore vectorStore;
    
    @Autowired
    @Qualifier("mcpClient")
    private ChatClient chatClient;
    
    /**
     * 获取个性化推荐（基于查询历史和上下文）
     * 
     * @param userQuery 用户查询
     * @param topK 返回数量
     * @return 推荐列表
     */
    public List<Recommendation> getPersonalizedRecommendations(String userQuery, int topK) {
        List<Recommendation> recommendations = new ArrayList<>();
        
        try {
            if (vectorStore != null) {
                // 1. 使用向量相似度搜索相关内容
                List<Document> similarDocs = vectorStore.similaritySearch(
                        SearchRequest.builder()
                                .query(userQuery)  // 使用构建器设置query
                                .topK(topK * 2)
                                .build()
                );
                
                // 2. 转换为推荐结果
                if (similarDocs != null) {
                    for (Document doc : similarDocs) {
                        Recommendation rec = new Recommendation();
                        rec.setItemId(doc.getId());
                        rec.setItemName(extractTitle(doc.getText()));
                        rec.setItemType(extractType(doc.getMetadata()));
                        rec.setScore(calculateScore(doc));
                        rec.setReason("基于向量相似度匹配");
                        rec.setSource("content_based");
                        rec.setMetadata(doc.getMetadata());

                        recommendations.add(rec);
                    }
                }

                // 3. 按分数排序并取前 topK
                recommendations = recommendations.stream()
                    .sorted((a, b) -> Double.compare(b.getScore(), a.getScore()))
                    .limit(topK)
                    .collect(Collectors.toList());
                    
                log.info("Generated {} recommendations for query: {}", recommendations.size(), userQuery);
            } else {
                log.warn("VectorStore not available, using fallback recommendations");
                recommendations = getFallbackRecommendations(topK);
            }
        } catch (Exception e) {
            log.error("Error generating recommendations", e);
            recommendations = getFallbackRecommendations(topK);
        }
        
        return recommendations;
    }
    
    /**
     * 获取基于上下文的智能推荐（流式响应）
     * 
     * @param context 上下文信息
     * @param userQuery 用户查询
     * @return 流式推荐文本
     */
    public Flux<String> getContextualRecommendations(String context, String userQuery) {
        return chatClient.prompt()
            .system("""
                你是星露谷物语的智能推荐助手。
                
                职责：
                1. 分析用户的查询和上下文
                2. 推荐 3-5 个相关的游戏内容、策略或物品
                3. 为每个推荐提供简短的理由
                
                输出格式：
                📌 推荐1：[名称]
                   理由：[为什么推荐]
                
                📌 推荐2：[名称]
                   理由：[为什么推荐]
                
                保持推荐简洁、实用、有针对性。
            """)
            .user("上下文：" + context + "\n用户问题：" + userQuery)
            .stream()
            .content()
            .doOnNext(chunk -> log.debug("Recommendation chunk: {}", chunk))
            .onErrorResume(e -> {
                log.error("Error generating contextual recommendations", e);
                return Flux.just("抱歉，推荐生成出现问题：" + e.getMessage());
            });
    }
    
    /**
     * 获取相似内容推荐
     * 
     * @param itemId 当前物品ID
     * @param itemType 物品类型
     * @param topK 返回数量
     * @return 相似内容列表
     */
    public List<Recommendation> getSimilarItems(String itemId, String itemType, int topK) {
        String query = String.format("与 %s 类似的 %s", itemId, itemType);
        return getPersonalizedRecommendations(query, topK);
    }
    
    /**
     * 获取智能策略建议
     * 
     * @param scenario 场景描述
     * @return 流式策略建议
     */
    public Flux<String> getStrategyRecommendations(String scenario) {
        return chatClient.prompt()
            .system("""
                你是星露谷物语的策略专家。
                
                任务：根据用户提供的场景，给出优化建议。
                
                分析维度：
                1. 经济效益
                2. 时间效率
                3. 风险评估
                4. 长期规划
                
                提供 3-5 个具体的可执行建议。
            """)
            .user("场景：" + scenario)
            .stream()
            .content();
    }
    
    /**
     * 从文档内容提取标题
     */
    private String extractTitle(String content) {
        if (content == null || content.isEmpty()) {
            return "未知项目";
        }
        // 提取第一行或前50个字符作为标题
        String[] lines = content.split("\n");
        String firstLine = lines[0].trim();
        return firstLine.length() > 50 ? firstLine.substring(0, 50) + "..." : firstLine;
    }
    
    /**
     * 从元数据提取类型
     */
    private String extractType(Map<String, Object> metadata) {
        if (metadata != null && metadata.containsKey("type")) {
            return metadata.get("type").toString();
        }
        return "general";
    }
    
    /**
     * 计算推荐分数
     */
    private Double calculateScore(Document doc) {
        // 可以基于相似度分数、新鲜度等因素计算
        // 这里简化处理，使用随机分数作为示例
        return 0.7 + (Math.random() * 0.3);
    }
    
    /**
     * 获取备用推荐（当向量存储不可用时）
     */
    private List<Recommendation> getFallbackRecommendations(int topK) {
        List<Recommendation> fallback = new ArrayList<>();
        
        // 添加一些默认的热门推荐
        String[] popularItems = {
            "草莓种子", "蓝莓种子", "蔓越莓种子", "南瓜种子", "古代种子"
        };
        
        for (int i = 0; i < Math.min(topK, popularItems.length); i++) {
            Recommendation rec = new Recommendation();
            rec.setItemId("fallback_" + i);
            rec.setItemName(popularItems[i]);
            rec.setItemType("crop");
            rec.setScore(0.8 - (i * 0.1));
            rec.setReason("热门推荐");
            rec.setSource("fallback");
            fallback.add(rec);
        }
        
        return fallback;
    }
    
    /**
     * 推荐多样性增强
     * 避免推荐结果过于相似
     */
    public List<Recommendation> diversifyRecommendations(List<Recommendation> recommendations) {
        if (recommendations.size() <= 3) {
            return recommendations;
        }
        
        List<Recommendation> diversified = new ArrayList<>();
        Set<String> usedTypes = new HashSet<>();
        
        // 优先选择不同类型的推荐
        for (Recommendation rec : recommendations) {
            if (!usedTypes.contains(rec.getItemType())) {
                diversified.add(rec);
                usedTypes.add(rec.getItemType());
            }
        }
        
        // 填充剩余的位置
        for (Recommendation rec : recommendations) {
            if (!diversified.contains(rec) && diversified.size() < recommendations.size()) {
                diversified.add(rec);
            }
        }
        
        return diversified;
    }
}
