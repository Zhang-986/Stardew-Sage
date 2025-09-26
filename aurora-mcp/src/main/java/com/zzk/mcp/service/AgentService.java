package com.zzk.mcp.service;

import reactor.core.publisher.Flux;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.stereotype.Service;

import java.time.Duration;

@Service
public class AgentService {
    private final ChatClient chatClient;

    public AgentService(ChatClient.Builder chatClientBuilder, ToolCallbackProvider tools) {
        String systemPrompt = """
                你作为星露谷Agent，请严格按以下步骤为用户提供趣味体验：
                
                ### 步骤1：调用工具'getTodayBirthday'获取星露谷今天生日信息
                
                ### 步骤2：根据人物JSON信息,结合他(她)的具体day,简短讲解关联
                """;

        this.chatClient = chatClientBuilder
                .defaultSystem(systemPrompt)
                .defaultToolCallbacks(tools)
                .build();
    }

    public Flux<String> getBirthdayInfoStream() {
        return chatClient.prompt()
                .user("获取今日生日信息")
                .stream()
                .content()
                .onErrorResume(e -> Flux.just("data: [ERROR] " + e.getMessage() + "\n\n"))
                .timeout(Duration.ofMinutes(10)); // 设置更长超时
    }
}