package com.zzk.auroamcpserver.config;

import com.zzk.auroamcpserver.service.WeatherService;
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
    public ToolCallbackProvider weatherTools(WeatherService weatherService){
        return MethodToolCallbackProvider.builder().
                toolObjects(weatherService)
                .build();
    }



}