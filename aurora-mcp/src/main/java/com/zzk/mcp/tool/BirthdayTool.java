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

    @Tool(description = "获取当天人物的生日情况")
    public String getTodayBirthday() {
        return birthdayService.getTodayBirthday();
    }
}
