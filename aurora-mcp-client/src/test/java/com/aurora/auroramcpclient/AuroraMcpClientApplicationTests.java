package com.aurora.auroramcpclient;

import io.modelcontextprotocol.client.McpSyncClient;
import io.modelcontextprotocol.spec.McpSchema;
import lombok.extern.slf4j.Slf4j;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import java.util.List;
import java.util.Map;

@Slf4j
@SpringBootTest
class DirectToolCallTest {

    @Autowired
    private List<McpSyncClient> mcpSyncClients;

    @Test
    void directCallWeatherTool() {
        if (mcpSyncClients == null || mcpSyncClients.isEmpty()) {
            log.error("❌ No available MCP clients found");
            return;
        }

        // 目标工具信息
        String targetToolName = "getWeather";
        String testCity = "北京"; // 测试参数

        mcpSyncClients.forEach(client -> {
            try {
                // 1. 验证工具是否存在
                McpSchema.ListToolsResult toolsResult = client.listTools();
                boolean toolExists = toolsResult.tools().stream()
                        .anyMatch(tool -> tool.name().equals(targetToolName));

                if (!toolExists) {
                    log.warn("⚠️ Tool {} not found in client {}", targetToolName, client.getClientInfo());
                    return;
                }

                // 2. 构建工具调用请求
                McpSchema.CallToolRequest request = new McpSchema.CallToolRequest(
                        targetToolName,
                        Map.of("city", testCity) // 参数必须匹配@ToolParam定义
                );

                // 3. 执行调用
                McpSchema.CallToolResult callToolResult = client.callTool(request);

                callToolResult.content().stream().forEach(System.out::println);

            } catch (Exception e) {
                log.error("🚨 Exception during tool call", e);
            }
        });
    }
}