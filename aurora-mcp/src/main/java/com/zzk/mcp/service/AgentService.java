package com.zzk.mcp.service;

import com.zzk.mcp.prompt.PeoplePrompt;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Flux;

import java.time.Duration;

@Service
public class AgentService {
    private final ChatClient chatClient;
    private final ToolCallbackProvider tools;

    // 不再设置默认 Prompt
    public AgentService(ChatClient.Builder chatClientBuilder, ToolCallbackProvider tools) {
        this.chatClient = chatClientBuilder
                .defaultToolCallbacks(tools) // 只保留默认工具回调
                .build();
        this.tools = tools;
    }

    // 生日信息流（使用生日 Prompt）
    public Flux<String> getBirthdayInfoStream() {
        return chatClient
                .prompt()
                .system(PeoplePrompt.BIRTHDAY_PROMPT) // 动态设置 Prompt
                .user("获取今日生日信息")
                .stream()
                .content()
                .onErrorResume(e -> Flux.just("data: [ERROR] " + e.getMessage() + "\n\n"))
                .timeout(Duration.ofMinutes(10));
    }

    // 任务信息流（使用任务 Prompt）
    public Flux<String> getMissionInfoStream() {
        return chatClient
                .prompt()
                .system(PeoplePrompt.MISSION_PROMPT) // 动态设置任务 Prompt
                .user("获取一个随机任务")
                .stream()
                .content()
                .onErrorResume(e -> Flux.just("data: [ERROR] " + e.getMessage() + "\n\n"))
                .timeout(Duration.ofMinutes(10));
    }
}