package com.zzk.mcp.tool;

import com.zzk.mcp.service.MissionService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class MissionTool {
    @Autowired
    private MissionService missionService;

    @Tool(name = "getMission",description = "获取当日任务")
    public String getMission() {
        return missionService.getRandomMission();
    }
}
