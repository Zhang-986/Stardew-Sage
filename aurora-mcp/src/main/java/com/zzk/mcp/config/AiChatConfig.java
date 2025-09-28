package com.zzk.mcp.config;

import lombok.RequiredArgsConstructor;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.chat.client.advisor.MessageChatMemoryAdvisor;
import org.springframework.ai.chat.client.advisor.SimpleLoggerAdvisor;
import org.springframework.ai.chat.client.advisor.vectorstore.QuestionAnswerAdvisor;
import org.springframework.ai.chat.memory.ChatMemory;
import org.springframework.ai.openai.OpenAiChatModel;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.ai.vectorstore.VectorStore;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
@RequiredArgsConstructor
@Configuration
public class AiChatConfig {
    @Bean
    @Qualifier("ragClient")
    public ChatClient ragClient(
            OpenAiChatModel chatModel,
            ChatMemory chatMemory,
            VectorStore vectorStore) {
        return ChatClient.builder(chatModel)
                .defaultSystem("请根据上下文回答,数据库里没有的,就不要擅自随便回答")
                .defaultAdvisors(
                        new MessageChatMemoryAdvisor(chatMemory),
                        new SimpleLoggerAdvisor(),
                        new QuestionAnswerAdvisor(vectorStore)
                ).build();
    }

    // 纯净的MCP服务调用ChatModel
    @Bean
    @Qualifier("mcpClient")
    public ChatClient mcpClient(OpenAiChatModel chatModel, ToolCallbackProvider tools) {
        return ChatClient.builder(chatModel)
                .defaultToolCallbacks(tools)
                .build();
    }
}
