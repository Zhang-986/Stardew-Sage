package com.zzk.mcp.service;

import com.alibaba.fastjson2.JSON;
import com.zzk.mcp.mapper.StardewMissionMapper;
import com.zzk.mcp.model.StardewMissionEntity;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

@RequiredArgsConstructor
@Service
public class MissionService {

    private static final Logger log = LoggerFactory.getLogger(MissionService.class);
    private final StardewMissionMapper stardewMissionMapper;

    public String getRandomMission() {
        int min = 1;
        int max = 165;

        int randomNumber = (int) (Math.random() * (max - min + 1)) + min;
        StardewMissionEntity stardewMissionEntity = stardewMissionMapper.selectById(randomNumber);
        return JSON.toJSONString(stardewMissionEntity);
    }
}
