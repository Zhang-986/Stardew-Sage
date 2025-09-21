package com.zzk.auroamcpserver.service;

import org.springframework.ai.tool.annotation.Tool;
import org.springframework.ai.tool.annotation.ToolParam;
import org.springframework.stereotype.Service;

@Service
public class WeatherService {

    @Tool(description = "获取城市天气")
    public String getWeather(@ToolParam(description = "城市") String city) {
        // 这里可以调用实际的天气API获取天气信息
        return "当前" + city + "的天气是晴天，温度25摄氏度。";
    }
}
