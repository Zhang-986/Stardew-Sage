package com.zzk.mcp;

import com.alibaba.fastjson2.JSON;
import com.zzk.mcp.model.StardewAnimalEntity;
import com.zzk.mcp.service.ResourceService;
import com.zzk.mcp.util.MetadataUtils;
import org.junit.jupiter.api.Test;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.document.Document;
import org.springframework.ai.vectorstore.SearchRequest;
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
    public void testVectorStore(){
        SearchRequest request = SearchRequest.builder()
                .query("800以上的动物")
                .topK(5)
                .similarityThreshold(0.6)
                .build();
        List<Document> documentList = vectorStore.similaritySearch(request);
        if(documentList != null){
            for (Document document : documentList) {
                System.out.println(document.getId());
                System.out.println(document.getScore());
                System.out.println(document.getText());
            }
        }
    }
    @Test
    void reloadMetadata() {
        System.out.println(ragChatClient.prompt()
                .user("星露谷性价比高的动物推荐")
                .call()
                .content());
    }

    @Test
    void search() {
        vectorStore.similaritySearch("牛")
                .stream().forEach(System.out::println);
    }
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
