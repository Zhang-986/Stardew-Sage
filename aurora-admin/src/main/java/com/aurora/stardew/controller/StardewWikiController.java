package com.aurora.stardew.controller;

import com.aurora.common.annotation.Log;
import com.aurora.common.core.controller.BaseController;
import com.aurora.common.core.domain.AjaxResult;
import com.aurora.common.core.page.TableDataInfo;
import com.aurora.common.enums.BusinessType;
import com.aurora.common.utils.poi.ExcelUtil;
import com.aurora.stardew.domain.StardewWiki;
import com.aurora.stardew.service.IStardewWikiService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletResponse;
import java.util.List;

/**
 * 星露谷Controller
 *
 * @author ruoyi
 * @date 2025-09-19
 */
@RestController
@RequestMapping("/stardew/wiki")
public class StardewWikiController extends BaseController {
    @Autowired
    private IStardewWikiService stardewWikiService;

    /**
     * 查询星露谷列表
     */
    @PreAuthorize("@ss.hasPermi('stardew:wiki:list')")
    @GetMapping("/list")
    public TableDataInfo list(StardewWiki stardewWiki) {
        startPage();
        List<StardewWiki> list = stardewWikiService.selectStardewWikiList(stardewWiki);
        return getDataTable(list);
    }

    /**
     * 导出星露谷列表
     */
    @PreAuthorize("@ss.hasPermi('stardew:wiki:export')")
    @Log(title = "星露谷", businessType = BusinessType.EXPORT)
    @PostMapping("/export")
    public void export(HttpServletResponse response, StardewWiki stardewWiki) {
        List<StardewWiki> list = stardewWikiService.selectStardewWikiList(stardewWiki);
        ExcelUtil<StardewWiki> util = new ExcelUtil<StardewWiki>(StardewWiki.class);
        util.exportExcel(response, list, "星露谷数据");
    }

    /**
     * 获取星露谷详细信息
     */
    @PreAuthorize("@ss.hasPermi('stardew:wiki:query')")
    @GetMapping(value = "/{id}")
    public AjaxResult getInfo(@PathVariable("id") Integer id) {
        return success(stardewWikiService.selectStardewWikiById(id));
    }

    /**
     * 新增星露谷
     */
    @PreAuthorize("@ss.hasPermi('stardew:wiki:add')")
    @Log(title = "星露谷", businessType = BusinessType.INSERT)
    @PostMapping
    public AjaxResult add(@RequestBody StardewWiki stardewWiki) {
        return toAjax(stardewWikiService.insertStardewWiki(stardewWiki));
    }

    /**
     * 修改星露谷
     */
    @PreAuthorize("@ss.hasPermi('stardew:wiki:edit')")
    @Log(title = "星露谷", businessType = BusinessType.UPDATE)
    @PutMapping
    public AjaxResult edit(@RequestBody StardewWiki stardewWiki) {
        return toAjax(stardewWikiService.updateStardewWiki(stardewWiki));
    }

    /**
     * 删除星露谷
     */
    @PreAuthorize("@ss.hasPermi('stardew:wiki:remove')")
    @Log(title = "星露谷", businessType = BusinessType.DELETE)
    @DeleteMapping("/{ids}")
    public AjaxResult remove(@PathVariable Integer[] ids) {
        return toAjax(stardewWikiService.deleteStardewWikiByIds(ids));
    }
}
