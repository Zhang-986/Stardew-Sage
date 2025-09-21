package com.aurora.stardew.service.impl;

import com.aurora.stardew.domain.StardewWiki;
import com.aurora.stardew.mapper.StardewWikiMapper;
import com.aurora.stardew.service.IStardewWikiService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 星露谷Service业务层处理
 * 
 * @author ruoyi
 * @date 2025-09-19
 */
@Service
public class StardewWikiServiceImpl implements IStardewWikiService 
{
    @Autowired
    private StardewWikiMapper stardewWikiMapper;

    /**
     * 查询星露谷
     * 
     * @param id 星露谷主键
     * @return 星露谷
     */
    @Override
    public StardewWiki selectStardewWikiById(Integer id)
    {
        return stardewWikiMapper.selectStardewWikiById(id);
    }

    /**
     * 查询星露谷列表
     * 
     * @param stardewWiki 星露谷
     * @return 星露谷
     */
    @Override
    public List<StardewWiki> selectStardewWikiList(StardewWiki stardewWiki)
    {
        return stardewWikiMapper.selectStardewWikiList(stardewWiki);
    }

    /**
     * 新增星露谷
     * 
     * @param stardewWiki 星露谷
     * @return 结果
     */
    @Override
    public int insertStardewWiki(StardewWiki stardewWiki)
    {
        return stardewWikiMapper.insertStardewWiki(stardewWiki);
    }

    /**
     * 修改星露谷
     * 
     * @param stardewWiki 星露谷
     * @return 结果
     */
    @Override
    public int updateStardewWiki(StardewWiki stardewWiki)
    {
        return stardewWikiMapper.updateStardewWiki(stardewWiki);
    }

    /**
     * 批量删除星露谷
     * 
     * @param ids 需要删除的星露谷主键
     * @return 结果
     */
    @Override
    public int deleteStardewWikiByIds(Integer[] ids)
    {
        return stardewWikiMapper.deleteStardewWikiByIds(ids);
    }

    /**
     * 删除星露谷信息
     * 
     * @param id 星露谷主键
     * @return 结果
     */
    @Override
    public int deleteStardewWikiById(Integer id)
    {
        return stardewWikiMapper.deleteStardewWikiById(id);
    }
}
