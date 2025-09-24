package com.zzk.mcp.controller;

import com.zzk.mcp.service.AgentService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * 获取天气信息Controller
 */
@RestController
@RequestMapping("/api/agent")
public class AgentController {
 
    @Autowired
    private AgentService agentService;
 
    @GetMapping("getDressingAdvice/{userInput}")
    public String getDressingAdvice(@PathVariable String userInput) {
        return agentService.getDressingAdvice(userInput);
    }
}