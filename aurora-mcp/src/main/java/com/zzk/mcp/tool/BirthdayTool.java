package com.zzk.mcp.tool;

import com.zzk.mcp.service.BirthdayService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class BirthdayTool {

    @Autowired
    private BirthdayService birthdayService;

    @Tool(name = "getTodayBirthday",description = "获取今日星露谷NPC玩家生日情况")
    public String getTodayBirthday() {
        return birthdayService.getTodayBirthday();
    }
}
