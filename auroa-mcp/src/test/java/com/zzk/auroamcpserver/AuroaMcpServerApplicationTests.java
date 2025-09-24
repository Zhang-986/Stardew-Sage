package com.zzk.auroamcpserver;

import com.zzk.mcp.service.AIService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

@SpringBootTest
class AuroaMcpServerApplicationTests {
    @Autowired
    private AIService aiService;
    @Test
    void contextLoads() {
        System.out.println(aiService.sendMessage("hello?"));
    }

}
