package com.zzk.mcp.controller;

import com.zzk.mcp.model.Recommendation;
import com.zzk.mcp.service.RecommendationService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import reactor.core.publisher.Flux;

import java.util.List;

/**
 * 智能推荐控制器
 * 提供多种推荐功能的 REST API
 * 
 * @author Stardew Sage Team
 */
@Slf4j
@RestController
@RequestMapping("/api/recommendation")
public class RecommendationController {
    
    @Autowired
    private RecommendationService recommendationService;
    
    /**
     * 获取个性化推荐
     * 
     * @param query 用户查询
     * @param topK 返回数量，默认 5
     * @return 推荐列表
     */
    @GetMapping("/personalized")
    public ResponseEntity<List<Recommendation>> getPersonalized(
        @RequestParam String query,
        @RequestParam(defaultValue = "5") int topK
    ) {
        log.info("Received personalized recommendation request: query={}, topK={}", query, topK);
        
        List<Recommendation> recommendations = 
            recommendationService.getPersonalizedRecommendations(query, topK);
        
        return ResponseEntity.ok(recommendations);
    }
    
    /**
     * 获取基于上下文的智能推荐（流式响应）
     * 
     * @param context 上下文信息
     * @param query 用户查询
     * @return 流式推荐文本
     */
    @GetMapping(value = "/contextual", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public Flux<String> getContextual(
        @RequestParam(required = false, defaultValue = "") String context,
        @RequestParam String query
    ) {
        log.info("Received contextual recommendation request: context={}, query={}", context, query);
        
        return recommendationService.getContextualRecommendations(context, query);
    }
    
    /**
     * 获取相似内容推荐
     * 
     * @param itemId 物品ID
     * @param itemType 物品类型
     * @param topK 返回数量，默认 5
     * @return 相似内容列表
     */
    @GetMapping("/similar")
    public ResponseEntity<List<Recommendation>> getSimilar(
        @RequestParam String itemId,
        @RequestParam String itemType,
        @RequestParam(defaultValue = "5") int topK
    ) {
        log.info("Received similar items request: itemId={}, itemType={}, topK={}", 
                 itemId, itemType, topK);
        
        List<Recommendation> recommendations = 
            recommendationService.getSimilarItems(itemId, itemType, topK);
        
        return ResponseEntity.ok(recommendations);
    }
    
    /**
     * 获取策略建议（流式响应）
     * 
     * @param scenario 场景描述
     * @return 流式策略建议
     */
    @GetMapping(value = "/strategy", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public Flux<String> getStrategy(@RequestParam String scenario) {
        log.info("Received strategy recommendation request: scenario={}", scenario);
        
        return recommendationService.getStrategyRecommendations(scenario);
    }
    
    /**
     * 获取多样化的推荐（避免推荐过于相似）
     * 
     * @param query 用户查询
     * @param topK 返回数量，默认 5
     * @return 多样化的推荐列表
     */
    @GetMapping("/diversified")
    public ResponseEntity<List<Recommendation>> getDiversified(
        @RequestParam String query,
        @RequestParam(defaultValue = "5") int topK
    ) {
        log.info("Received diversified recommendation request: query={}, topK={}", query, topK);
        
        List<Recommendation> recommendations = 
            recommendationService.getPersonalizedRecommendations(query, topK * 2);
        
        List<Recommendation> diversified = 
            recommendationService.diversifyRecommendations(recommendations);
        
        // 限制返回数量
        if (diversified.size() > topK) {
            diversified = diversified.subList(0, topK);
        }
        
        return ResponseEntity.ok(diversified);
    }
}
