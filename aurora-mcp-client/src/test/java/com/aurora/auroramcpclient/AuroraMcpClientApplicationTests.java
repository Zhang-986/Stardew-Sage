package com.aurora.auroramcpclient;

import io.modelcontextprotocol.client.McpSyncClient;
import io.modelcontextprotocol.spec.McpSchema;
import lombok.extern.slf4j.Slf4j;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;
@Slf4j
@SpringBootTest
class AuroraMcpClientApplicationTests {
    @Autowired
    private List<McpSyncClient> mcpSyncClients;  // For sync client
    @Test
    void inspectSpecificTool() {
        if (mcpSyncClients == null || mcpSyncClients.isEmpty()) {
            return;
        }

        mcpSyncClients.forEach(client -> {
            try {
                List<McpSchema.Tool> tools = client.listTools().tools();

                tools.forEach(tool -> {
                    System.out.println("🔍 工具详情: " + tool.name());
                    System.out.println("   📝 描述: " + tool.description());

                    if (tool.inputSchema() != null) {
                        System.out.println("   🎯 输入参数: " + tool.inputSchema());
                    }


                    System.out.println("   -".repeat(20));
                });

            } catch (Exception e) {
                System.out.println("❌ 获取工具信息失败: " + e.getMessage());
            }
        });
    }
}
