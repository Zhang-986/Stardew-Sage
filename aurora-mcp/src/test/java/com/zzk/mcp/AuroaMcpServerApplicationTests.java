package com.zzk.mcp;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zzk.mcp.mapper.StardewCraftMapper;
import com.zzk.mcp.model.StardewCraftEntity;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import java.util.List;

@SpringBootTest(classes = AuroaMcpServerApplication.class)
class AuroaMcpServerApplicationTests {
   @Autowired
   private StardewCraftMapper stardewCraftMapper;
    @Test
    void contextLoads() {
        StardewCraftEntity stardewCraftEntity = new StardewCraftEntity();
        stardewCraftEntity.setCraftType("花花");
        stardewCraftEntity.setDataMode(1);
        stardewCraftEntity.setNameCh("花花aa");
        stardewCraftEntity.setSysCode("asd");
        stardewCraftMapper.insert(stardewCraftEntity);
    }

}
