package com.zzk.mcp.tool;

import com.zzk.mcp.service.AIService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.ai.tool.annotation.ToolParam;
import org.springframework.stereotype.Service;
@Slf4j
@Service
public class WeatherTool {

    private final AIService aIService;

    public WeatherTool(AIService aIService) {
        this.aIService = aIService;
    }

    @Tool(description = "获取城市天气")
    public String getWeather(@ToolParam(description = "城市") String city) {
        return city +"的天气是18°c，晴和多云";
    }
}
