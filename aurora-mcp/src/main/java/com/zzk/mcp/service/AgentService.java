package com.zzk.mcp.service;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.stereotype.Service;

/**
 * 星露谷用户信息助手
 */
@Service
public class AgentService {
    private final ChatClient chatClient;

    public AgentService(ChatClient.Builder chatClientBuilder, ToolCallbackProvider tools) {
        String systemPrompt = """
                你作为星露谷Agent，请严格按以下步骤为用户为用户提供趣味体验，
                
                ### 步骤1：根据tool工具调用数据库星露谷今天生日信息
                调用工具'getTodayBirthday'获取星露谷今天生日信息的人的信息
                
                ### 步骤2：匹配到任务数据项，讲解人物趣味故事。
                根据人物JSON信息,讲解今天人物的信息,风趣味
                
                """;

        this.chatClient = chatClientBuilder
                .defaultSystem(systemPrompt)
                .defaultToolCallbacks( tools)
                .build();
    }

    public String getBirthdayInfo() {
        return chatClient
                .prompt()
                .user("根据Prompt做一些东西")
                .call()
                .content();
    }
}