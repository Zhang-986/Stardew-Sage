package com.zzk.mcp.service;

import com.zzk.mcp.prompt.PeoplePrompt;
import reactor.core.publisher.Flux;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.stereotype.Service;

import java.time.Duration;

@Service
public class AgentService {
    private final ChatClient chatClient;

    public AgentService(ChatClient.Builder chatClientBuilder, ToolCallbackProvider tools) {
        this.chatClient =
                chatClientBuilder
                .defaultSystem(PeoplePrompt.BIRTHDAY_PROMPT)
                .defaultToolCallbacks(tools)
                .build();
    }

    public Flux<String> getBirthdayInfoStream() {
        return chatClient
                .prompt()
                .user("获取今日生日信息")
                .stream()
                .content()
                .onErrorResume(e -> Flux.just("data: [ERROR] " + e.getMessage() + "\n\n"))
                .timeout(Duration.ofMinutes(10)); // 设置更长超时
    }
}