package com.zzk.auroamcpserver.config;

import com.zzk.auroamcpserver.service.WeatherService;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.ai.tool.method.MethodToolCallbackProvider;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class ToolCallbackProviderRegister {
    @Bean
    public ToolCallbackProvider weatherTools(WeatherService weatherService){
        return MethodToolCallbackProvider.builder().
                toolObjects(weatherService)
                .build();
    }
}
