package com.zzk.mcp.controller;

import com.zzk.mcp.service.AgentService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import reactor.core.publisher.Flux;

@RestController
@RequestMapping("/api/agent")
public class AgentController {

    @Autowired
    private AgentService agentService;

    @GetMapping(value = "/getBirthInfoStream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public Flux<String> getBirthInfoStream() {
        return agentService.getBirthdayInfoStream();
    }
    @GetMapping(value = "/getMissionInfoStream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public Flux<String> getMissionInfoStream() {
        return agentService.getMissionInfoStream();
    }
    @GetMapping(value = "/getInfoDetail",produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public Flux<String> getInfoDetail(@RequestParam(value = "question", required = false)String question ) {
        return agentService.getInfoDetail(question);
    }
    @GetMapping(value = "/getInfoRagDetail",produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public Flux<String> getInfoRagDetail(@RequestParam(value = "question", required = false)String question ) {
        return agentService.getInfoRagDetail(question);
    }
}