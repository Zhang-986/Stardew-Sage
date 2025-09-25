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
        LambdaQueryWrapper<StardewCraftEntity> stardewCraftEntityLambdaQueryWrapper = new LambdaQueryWrapper<>();
        stardewCraftEntityLambdaQueryWrapper.isNotNull(StardewCraftEntity::getId);
        List<StardewCraftEntity> stardewCraftEntities = stardewCraftMapper.selectList(stardewCraftEntityLambdaQueryWrapper);
        for (StardewCraftEntity stardewCraftEntity : stardewCraftEntities) {
            System.out.println(stardewCraftEntity.getId());
            System.out.println(stardewCraftEntity.getSysCode());
            System.out.println(stardewCraftEntity.getNameCh());
            System.out.println(stardewCraftEntity.getCraftType());
            System.out.println(stardewCraftEntity.getMakePrice());
            System.out.println(stardewCraftEntity.getSellPrice());
            System.out.println(stardewCraftEntity.getSourceFrom());
            System.out.println(stardewCraftEntity.getRemark());
            System.out.println(stardewCraftEntity.getDataMode());
        }
    }

}
