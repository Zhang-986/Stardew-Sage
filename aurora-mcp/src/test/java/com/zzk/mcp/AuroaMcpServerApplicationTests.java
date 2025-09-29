package com.zzk.mcp;

import com.alibaba.fastjson2.JSON;
import com.zzk.mcp.model.StardewAnimalEntity;
import com.zzk.mcp.service.ResourceService;
import com.zzk.mcp.util.MetadataUtils;
import org.junit.jupiter.api.Test;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.document.Document;
import org.springframework.ai.vectorstore.redis.RedisVectorStore;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.boot.test.context.SpringBootTest;

import java.util.List;

@SpringBootTest(classes = AuroaMcpServerApplication.class)
class AuroaMcpServerApplicationTests {
    @Autowired
    @Qualifier("ragClient")
    private ChatClient ragChatClient; // 注入Builder便于动态调整

    @Autowired
    private ResourceService resourceService;

    @Autowired
    private RedisVectorStore vectorStore;

    @Test
    void contextLoads() {
        List<StardewAnimalEntity> animalInfo = resourceService.getAnimalInfo();
        List<Document> documentList = animalInfo.stream().map(
                item -> {
                    return Document.builder()
                            .metadata(MetadataUtils.convertToMetadata(item))
                            .text(JSON.toJSONString(item))
                            .build();
                }
        ).toList();
        vectorStore.add(documentList);
    }


}
