package com.zzk.mcp.config.properties;

import lombok.Getter;
import lombok.Setter;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;


@Setter
@Getter
@Component
@ConfigurationProperties(prefix = "ai.dashscope")
public class DashScopeProperties {
    private String apiKey;
    private String baseUrl;
    private String defaultModel = "qwen-plus";

}