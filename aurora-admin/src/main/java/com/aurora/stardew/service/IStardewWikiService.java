package com.aurora.stardew.service;

import com.aurora.stardew.domain.StardewWiki;

import java.util.List;

/**
 * 星露谷Service接口
 * 
 * @author ruoyi
 * @date 2025-09-19
 */
public interface IStardewWikiService 
{
    /**
     * 查询星露谷
     * 
     * @param id 星露谷主键
     * @return 星露谷
     */
    public StardewWiki selectStardewWikiById(Integer id);

    /**
     * 查询星露谷列表
     * 
     * @param stardewWiki 星露谷
     * @return 星露谷集合
     */
    public List<StardewWiki> selectStardewWikiList(StardewWiki stardewWiki);

    /**
     * 新增星露谷
     * 
     * @param stardewWiki 星露谷
     * @return 结果
     */
    public int insertStardewWiki(StardewWiki stardewWiki);

    /**
     * 修改星露谷
     * 
     * @param stardewWiki 星露谷
     * @return 结果
     */
    public int updateStardewWiki(StardewWiki stardewWiki);

    /**
     * 批量删除星露谷
     * 
     * @param ids 需要删除的星露谷主键集合
     * @return 结果
     */
    public int deleteStardewWikiByIds(Integer[] ids);

    /**
     * 删除星露谷信息
     * 
     * @param id 星露谷主键
     * @return 结果
     */
    public int deleteStardewWikiById(Integer id);
}
