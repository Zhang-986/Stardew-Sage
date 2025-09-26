package com.zzk.mcp.controller;

import com.zzk.mcp.service.AgentService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import reactor.core.publisher.Flux;

@RestController
@RequestMapping("/api/agent")
public class AgentController {

    @Autowired
    private AgentService agentService;

    @GetMapping(value = "/getDressingAdvice", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public Flux<String> getDressingAdviceStream() {
        return agentService.getBirthdayInfoStream();
    }
}