package com.zzk.mcp.service;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.stereotype.Service;

import java.time.LocalDate;
import java.time.format.DateTimeFormatter;

/**
 * 穿衣助手服务（直接调用MCP Server的@Tool方法）
 */
@Service
public class AgentService {
    private final ChatClient chatClient;

    public AgentService(ChatClient.Builder chatClientBuilder, ToolCallbackProvider tools) {
        String systemPrompt = """
                你作为穿衣助手Agent，请严格按以下步骤为用户推荐穿衣搭配：
                ### 步骤1：获取当前日期
                当前日期为{currentDate}。
                
                ### 步骤2：校验用户输入
                如果用户输入不包含城市名则提示用户'请输入城市名'。
                
                ### 步骤3：根据城市名获取天气信息
                调用工具'getWeather'获取天气信息，入参：城市名city（从用户输入中提取）。
                
                ### 步骤4：根据天气信息生成穿衣建议
                结合天气信息直接给出穿搭建议（无需再调用工具）。
                """;

        String currentDateStr = LocalDate.now().format(DateTimeFormatter.ofPattern("yyyy-MM-dd"));
        systemPrompt = systemPrompt.replace("{currentDate}", currentDateStr);

        this.chatClient = chatClientBuilder
                .defaultSystem(systemPrompt)
                .defaultToolCallbacks( tools)
                .build();
    }

    public String getDressingAdvice(String userInput) {
        return chatClient.prompt()
                .user(userInput)
                .call() // 自动触发getWeather工具调用
                .content();
    }
}