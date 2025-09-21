package com.aurora.dinner.controller;

import java.util.List;
import javax.servlet.http.HttpServletResponse;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import com.aurora.common.annotation.Log;
import com.aurora.common.core.controller.BaseController;
import com.aurora.common.core.domain.AjaxResult;
import com.aurora.common.enums.BusinessType;
import com.aurora.dinner.domain.LunchTable;
import com.aurora.dinner.service.ILunchTableService;
import com.aurora.common.utils.poi.ExcelUtil;
import com.aurora.common.core.page.TableDataInfo;

/**
 * 午市就餐-桌台Controller
 * 
 * @author ruoyi
 * @date 2025-09-19
 */
@RestController
@RequestMapping("/dinner/table")
public class LunchTableController extends BaseController
{
    @Autowired
    private ILunchTableService lunchTableService;

    /**
     * 查询午市就餐-桌台列表
     */
    @PreAuthorize("@ss.hasPermi('dinner:table:list')")
    @GetMapping("/list")
    public TableDataInfo list(LunchTable lunchTable)
    {
        startPage();
        List<LunchTable> list = lunchTableService.selectLunchTableList(lunchTable);
        return getDataTable(list);
    }

    /**
     * 导出午市就餐-桌台列表
     */
    @PreAuthorize("@ss.hasPermi('dinner:table:export')")
    @Log(title = "午市就餐-桌台", businessType = BusinessType.EXPORT)
    @PostMapping("/export")
    public void export(HttpServletResponse response, LunchTable lunchTable)
    {
        List<LunchTable> list = lunchTableService.selectLunchTableList(lunchTable);
        ExcelUtil<LunchTable> util = new ExcelUtil<LunchTable>(LunchTable.class);
        util.exportExcel(response, list, "午市就餐-桌台数据");
    }

    /**
     * 获取午市就餐-桌台详细信息
     */
    @PreAuthorize("@ss.hasPermi('dinner:table:query')")
    @GetMapping(value = "/{tableId}")
    public AjaxResult getInfo(@PathVariable("tableId") Long tableId)
    {
        return success(lunchTableService.selectLunchTableByTableId(tableId));
    }

    /**
     * 新增午市就餐-桌台
     */
    @PreAuthorize("@ss.hasPermi('dinner:table:add')")
    @Log(title = "午市就餐-桌台", businessType = BusinessType.INSERT)
    @PostMapping
    public AjaxResult add(@RequestBody LunchTable lunchTable)
    {
        return toAjax(lunchTableService.insertLunchTable(lunchTable));
    }

    /**
     * 修改午市就餐-桌台
     */
    @PreAuthorize("@ss.hasPermi('dinner:table:edit')")
    @Log(title = "午市就餐-桌台", businessType = BusinessType.UPDATE)
    @PutMapping
    public AjaxResult edit(@RequestBody LunchTable lunchTable)
    {
        return toAjax(lunchTableService.updateLunchTable(lunchTable));
    }

    /**
     * 删除午市就餐-桌台
     */
    @PreAuthorize("@ss.hasPermi('dinner:table:remove')")
    @Log(title = "午市就餐-桌台", businessType = BusinessType.DELETE)
	@DeleteMapping("/{tableIds}")
    public AjaxResult remove(@PathVariable Long[] tableIds)
    {
        return toAjax(lunchTableService.deleteLunchTableByTableIds(tableIds));
    }
}
