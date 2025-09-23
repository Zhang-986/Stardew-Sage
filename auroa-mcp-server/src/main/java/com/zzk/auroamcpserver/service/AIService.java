package com.zzk.auroamcpserver.service;

import com.openai.client.OpenAIClient;

import com.openai.models.ChatCompletion;
import com.openai.models.ChatCompletionCreateParams;
import com.zzk.auroamcpserver.ai.DashScopeProperties;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

@Service
public class AIService {

    private final OpenAIClient client;
    private final String defaultModel;

    public AIService(OpenAIClient client, DashScopeProperties properties) {
        this.client = client;
        // 确保使用配置中的默认模型，如果没有配置则使用"qwen-plus"
        this.defaultModel = properties.getDefaultModel() != null ?
                properties.getDefaultModel() : "qwen-plus";
    }

    /**
     * 发送简单消息并获取回复
     * @param message 用户消息
     * @return AI回复
     */
    public String sendMessage(String message) {
        return sendMessageWithModel(message, defaultModel);
    }

    /**
     * 带自定义模型的聊天
     * @param message 用户消息
     * @param model 指定模型
     * @return AI回复
     */
    public String sendMessageWithModel(String message, String model) {
        try {
            ChatCompletionCreateParams params = ChatCompletionCreateParams.builder()
                    .addUserMessage(message)
                    .model(model)
                    .build();

            ChatCompletion chatCompletion = client.chat().completions().create(params);
            return chatCompletion.choices().get(0).message().content().get();
        } catch (Exception e) {
            throw new RuntimeException("Failed to get AI response with model " + model, e);
        }
    }

}