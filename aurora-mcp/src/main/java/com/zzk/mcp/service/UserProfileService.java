package com.zzk.mcp.service;

import com.alibaba.fastjson2.JSON;
import com.zzk.mcp.model.UserProfile;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.HashMap;

@Slf4j
@Service
public class UserProfileService {

    @Autowired
    private ChatClient.Builder chatClientBuilder;

    @Autowired
    private RedisTemplate<String, String> redisTemplate;

    private static final String PROFILE_KEY_PREFIX = "user:profile:";

    /**
     * 获取用户画像
     */
    public UserProfile getUserProfile(String userId) {
        String key = PROFILE_KEY_PREFIX + userId;
        String json = redisTemplate.opsForValue().get(key);
        if (json != null) {
            return JSON.parseObject(json, UserProfile.class);
        }
        // 返回空画像
        UserProfile empty = new UserProfile();
        empty.setUserId(userId);
        empty.setTags(new ArrayList<>());
        empty.setPreferences(new HashMap<>());
        return empty;
    }

    /**
     * 根据新的对话内容更新用户画像
     */
    public UserProfile updateUserProfile(String userId, String chatContent) {
        UserProfile currentProfile = getUserProfile(userId);
        
        String systemPrompt = """
            你是一个用户画像专家。请根据【现有画像】和【最新对话】，更新用户的画像信息。
            请提取用户的职业、兴趣、偏好、性格特征等。
            
            请严格只返回 JSON 格式，不要包含 Markdown 标记，格式如下：
            {
                "tags": ["标签1", "标签2"],
                "summary": "用户的一句话简介",
                "preferences": {"key": "value"}
            }
            """;

        String userPrompt = String.format("""
            【现有画像】: %s
            【最新对话】: %s
            """, JSON.toJSONString(currentProfile), chatContent);

        try {
            ChatClient chatClient = chatClientBuilder.build();
            String resultJson = chatClient.prompt()
                    .system(systemPrompt)
                    .user(userPrompt)
                    .call()
                    .content();

            // 清理可能的 Markdown 标记
            resultJson = resultJson.replace("```json", "").replace("```", "").trim();

            UserProfile newInfo = JSON.parseObject(resultJson, UserProfile.class);
            
            if (newInfo.getTags() != null) currentProfile.setTags(newInfo.getTags());
            if (newInfo.getSummary() != null) currentProfile.setSummary(newInfo.getSummary());
            if (newInfo.getPreferences() != null) {
                currentProfile.getPreferences().putAll(newInfo.getPreferences());
            }
            currentProfile.setLastUpdate(LocalDateTime.now().toString());

            redisTemplate.opsForValue().set(PROFILE_KEY_PREFIX + userId, JSON.toJSONString(currentProfile));

        } catch (Exception e) {
            log.error("Failed to analyze user profile", e);
        }

        return currentProfile;
    }
}
