package com.aurora.dinner.service.impl;

import java.util.List;
import com.aurora.common.utils.DateUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import com.aurora.dinner.mapper.LunchTableMapper;
import com.aurora.dinner.domain.LunchTable;
import com.aurora.dinner.service.ILunchTableService;

/**
 * 午市就餐-桌台Service业务层处理
 * 
 * @author ruoyi
 * @date 2025-09-19
 */
@Service
public class LunchTableServiceImpl implements ILunchTableService 
{
    @Autowired
    private LunchTableMapper lunchTableMapper;

    /**
     * 查询午市就餐-桌台
     * 
     * @param tableId 午市就餐-桌台主键
     * @return 午市就餐-桌台
     */
    @Override
    public LunchTable selectLunchTableByTableId(Long tableId)
    {
        return lunchTableMapper.selectLunchTableByTableId(tableId);
    }

    /**
     * 查询午市就餐-桌台列表
     * 
     * @param lunchTable 午市就餐-桌台
     * @return 午市就餐-桌台
     */
    @Override
    public List<LunchTable> selectLunchTableList(LunchTable lunchTable)
    {
        return lunchTableMapper.selectLunchTableList(lunchTable);
    }

    /**
     * 新增午市就餐-桌台
     * 
     * @param lunchTable 午市就餐-桌台
     * @return 结果
     */
    @Override
    public int insertLunchTable(LunchTable lunchTable)
    {
        lunchTable.setCreateTime(DateUtils.getNowDate());
        return lunchTableMapper.insertLunchTable(lunchTable);
    }

    /**
     * 修改午市就餐-桌台
     * 
     * @param lunchTable 午市就餐-桌台
     * @return 结果
     */
    @Override
    public int updateLunchTable(LunchTable lunchTable)
    {
        lunchTable.setUpdateTime(DateUtils.getNowDate());
        return lunchTableMapper.updateLunchTable(lunchTable);
    }

    /**
     * 批量删除午市就餐-桌台
     * 
     * @param tableIds 需要删除的午市就餐-桌台主键
     * @return 结果
     */
    @Override
    public int deleteLunchTableByTableIds(Long[] tableIds)
    {
        return lunchTableMapper.deleteLunchTableByTableIds(tableIds);
    }

    /**
     * 删除午市就餐-桌台信息
     * 
     * @param tableId 午市就餐-桌台主键
     * @return 结果
     */
    @Override
    public int deleteLunchTableByTableId(Long tableId)
    {
        return lunchTableMapper.deleteLunchTableByTableId(tableId);
    }
}
