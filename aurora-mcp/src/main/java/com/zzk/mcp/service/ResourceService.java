package com.zzk.mcp.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zzk.mcp.mapper.StardewAnimalMapper;
import com.zzk.mcp.model.StardewAnimalEntity;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

@RequiredArgsConstructor
@Service
public class ResourceService {
    private final StardewAnimalMapper stardewAnimalMapper;

    public List<StardewAnimalEntity> getAnimalInfo(){
        LambdaQueryWrapper<StardewAnimalEntity> stardewAnimalEntityLambdaQueryWrapper = new LambdaQueryWrapper<>();
        stardewAnimalEntityLambdaQueryWrapper.select(StardewAnimalEntity::getNameCh,
                StardewAnimalEntity::getBuyPrice,
                StardewAnimalEntity::getSellPrice,
                StardewAnimalEntity::getNeedBuild,
                StardewAnimalEntity::getAnimalOutput,
                StardewAnimalEntity::getRemark);
        return stardewAnimalMapper.selectList(stardewAnimalEntityLambdaQueryWrapper);
    }
}
