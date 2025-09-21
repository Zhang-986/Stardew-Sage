package com.aurora.dinner.mapper;

import java.util.List;
import com.aurora.dinner.domain.LunchTable;

/**
 * 午市就餐-桌台Mapper接口
 * 
 * @author ruoyi
 * @date 2025-09-19
 */
public interface LunchTableMapper 
{
    /**
     * 查询午市就餐-桌台
     * 
     * @param tableId 午市就餐-桌台主键
     * @return 午市就餐-桌台
     */
    public LunchTable selectLunchTableByTableId(Long tableId);

    /**
     * 查询午市就餐-桌台列表
     * 
     * @param lunchTable 午市就餐-桌台
     * @return 午市就餐-桌台集合
     */
    public List<LunchTable> selectLunchTableList(LunchTable lunchTable);

    /**
     * 新增午市就餐-桌台
     * 
     * @param lunchTable 午市就餐-桌台
     * @return 结果
     */
    public int insertLunchTable(LunchTable lunchTable);

    /**
     * 修改午市就餐-桌台
     * 
     * @param lunchTable 午市就餐-桌台
     * @return 结果
     */
    public int updateLunchTable(LunchTable lunchTable);

    /**
     * 删除午市就餐-桌台
     * 
     * @param tableId 午市就餐-桌台主键
     * @return 结果
     */
    public int deleteLunchTableByTableId(Long tableId);

    /**
     * 批量删除午市就餐-桌台
     * 
     * @param tableIds 需要删除的数据主键集合
     * @return 结果
     */
    public int deleteLunchTableByTableIds(Long[] tableIds);
}
