package com.zzk.mcp.controller;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.core.io.ByteArrayResource;
import org.springframework.util.MimeType;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.multipart.MultipartFile;
import reactor.core.publisher.Flux;

import java.io.IOException;
import java.time.Duration;

@RestController
@RequestMapping("/api/image")
public class ImageController {

    @Autowired
    @Qualifier("visionAnalysisClient")
    private ChatClient visionAnalysisClient; // 注入多模态客户端

    /**
     * 分析前端直接上传的图片文件
     *
     * @param file     前端上传的图片文件
     * @param question 用户提问（可选）
     */
    @PostMapping("/analyze")
    public Flux<String> analyzeImage(@RequestParam("file") MultipartFile file,
                                     @RequestParam(required = false) String question) {

        // 基础验证
        if (file.isEmpty()) {
            return Flux.just("错误：请上传有效的图片文件");
        }

        if (!file.getContentType().startsWith("image/")) {
            return Flux.just("错误：仅支持图片文件（JPEG/PNG等）");
        }

        if (file.getSize() > 5 * 1024 * 1024) {
            return Flux.just("错误：图片大小不能超过5MB");
        }
        try {
            // 创建Media对象
            MimeType mimeType = MimeType.valueOf(file.getContentType());
            ByteArrayResource resource = new ByteArrayResource(file.getBytes());

            return visionAnalysisClient
                    .prompt()
                    .system("请联系星露谷物语进行图片分析")
                    .user(u -> u
                            .text(question != null ? question : "请描述这张图片")
                            .media(mimeType,resource)  // 使用Media对象
                    )
                    .stream()
                    .content()
                    .onErrorResume(e -> Flux.just("data: [ERROR] " + e.getMessage() + "\n\n"))
                    .timeout(Duration.ofMinutes(10));
        } catch (IOException e) {
            return Flux.just("data: [ERROR] 文件处理失败: " + e.getMessage() + "\n\n");
        }

    }
}
