package com.zzk.mcp.config;

import com.zzk.mcp.tool.BirthdayTool;
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
    public ToolCallbackProvider weatherTools(BirthdayTool birthdayTool){
        return MethodToolCallbackProvider.builder().
                toolObjects(birthdayTool)
                .build();
    }

}