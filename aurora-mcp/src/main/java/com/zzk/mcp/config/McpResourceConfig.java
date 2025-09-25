package com.zzk.mcp.config;

import com.openai.client.OpenAIClient;
import com.openai.client.okhttp.OpenAIOkHttpClient;
import com.zzk.mcp.config.properties.DashScopeProperties;
import com.zzk.mcp.tool.WeatherTool;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.ai.tool.method.MethodToolCallbackProvider;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * MCP Resources 配置类
 * 专门负责注册和管理所有资源
 */
@Configuration
public class McpResourceConfig {


    @Bean
    public ToolCallbackProvider weatherTools(WeatherTool weatherTool){
        return MethodToolCallbackProvider.builder().
                toolObjects(weatherTool)
                .build();
    }


    @Configuration
    public static class AIClientConfig {

        private final DashScopeProperties properties;

        public AIClientConfig(DashScopeProperties properties) {
            this.properties = properties;
        }

        @Bean
        public OpenAIClient openAIClient() {
            return OpenAIOkHttpClient.builder()
                    .apiKey(properties.getApiKey())
                    .baseUrl(properties.getBaseUrl())
                    .build();
        }
    }
}