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
                .defaultAdvisors(new SimpleLoggerAdvisor())
                .build();
    }

    @Bean
    @Qualifier("visionAnalysisClient")
    public ChatClient visionAnalysisClient(
            OpenAiChatModel chatModel, // 必须是支持多模态的ChatModel
            ChatMemory chatMemory) {

        return ChatClient.builder(chatModel)
                .defaultSystem("""
                    你是一个星露谷物语图像分析专家，请严格遵守：
                    1. 只分析用户提供的图像内容
                    2. 结合游戏知识回答（优先使用RAG检索结果）
                    3. 若图像与游戏无关，回答："请提供星露谷物语相关图片"
                    """)
                .defaultAdvisors(
                        new MessageChatMemoryAdvisor(chatMemory),
                        new SimpleLoggerAdvisor()
                )
                .build();
    }
}
