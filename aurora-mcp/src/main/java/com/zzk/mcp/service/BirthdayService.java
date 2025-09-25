package com.zzk.mcp.service;

import com.alibaba.fastjson2.JSON;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zzk.mcp.mapper.StardewPeopleFavoriteMapper;
import com.zzk.mcp.mapper.StardewPeopleMapper;
import com.zzk.mcp.model.StardewPeopleEntity;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Calendar;
import java.util.List;
@RequiredArgsConstructor
@Service
public class BirthdayService {


    private final StardewPeopleMapper stardewPeopleMapper;


    public String getTodayBirthday() {
        // 获取当前月份和日期
        Calendar calendar = Calendar.getInstance();
        int month = calendar.get(Calendar.MONTH) + 1;
        int day = calendar.get(Calendar.DATE);

        // 将现实月份转换为星露谷季节
        String season = convertToStardewSeason(month);

        // 构建查询条件
        LambdaQueryWrapper<StardewPeopleEntity> queryWrapper = new LambdaQueryWrapper<>();
        queryWrapper.like(StardewPeopleEntity::getBirthday, season+day);

        // 执行查询
        List<StardewPeopleEntity> birthdayPeople = stardewPeopleMapper.selectList(queryWrapper);

        // 处理结果
        if (birthdayPeople.isEmpty()) {
            return "今天在星露谷没有NPC过生日";
        }

        return JSON.toJSONString(birthdayPeople);
    }

    /**
     * 将现实月份转换为星露谷季节
     */
    private String convertToStardewSeason(int month) {
        switch (month) {
            case 1:
            case 2:
            case 3:
                return "春季";
            case 4:
            case 5:
            case 6:
                return "夏季";
            case 7:
            case 8:
            case 9:
                return "秋季";
            case 10:
            case 11:
            case 12:
                return "冬冬";
            default:
                return "未知";
        }
    }
}
