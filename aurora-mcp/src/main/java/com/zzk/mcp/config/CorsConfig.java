package com.zzk.mcp.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class CorsConfig implements WebMvcConfigurer {

    @Override
    public void addCorsMappings(CorsRegistry registry) {
        registry.addMapping("/**")
                .allowedOrigins("*")
                .allowedMethods("*")
                .allowedHeaders("*")
                .exposedHeaders("Content-Disposition") // 暴露特殊头
                .allowCredentials(false);
        
        // 特别为SSE流式接口配置
        registry.addMapping("/api/agent/**")
                .allowedOrigins("*")
                .allowedMethods("GET")
                .allowedHeaders("Cache-Control", "Content-Type")
                .exposedHeaders("text/event-stream")
                .allowCredentials(false);
    }
}