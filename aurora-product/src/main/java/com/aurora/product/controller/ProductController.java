package com.aurora.product.controller;

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
import com.aurora.product.domain.Product;
import com.aurora.product.service.IProductService;
import com.aurora.common.utils.poi.ExcelUtil;
import com.aurora.common.core.page.TableDataInfo;

/**
 * 商品库存管理Controller
 * 
 * @author aurora
 * @date 2025-08-04
 */
@RestController
@RequestMapping("/product/stock")
public class ProductController extends BaseController
{
    @Autowired
    private IProductService productService;

    /**
     * 查询商品库存管理列表
     */
    @PreAuthorize("@ss.hasPermi('product:stock:list')")
    @GetMapping("/list")
    public TableDataInfo list(Product product)
    {
        startPage();
        List<Product> list = productService.selectProductList(product);
        return getDataTable(list);
    }

    /**
     * 导出商品库存管理列表
     */
    @PreAuthorize("@ss.hasPermi('product:stock:export')")
    @Log(title = "商品库存管理", businessType = BusinessType.EXPORT)
    @PostMapping("/export")
    public void export(HttpServletResponse response, Product product)
    {
        List<Product> list = productService.selectProductList(product);
        ExcelUtil<Product> util = new ExcelUtil<Product>(Product.class);
        util.exportExcel(response, list, "商品库存管理数据");
    }

    /**
     * 获取商品库存管理详细信息
     */
    @PreAuthorize("@ss.hasPermi('product:stock:query')")
    @GetMapping(value = "/{id}")
    public AjaxResult getInfo(@PathVariable("id") Long id)
    {
        return success(productService.selectProductById(id));
    }

    /**
     * 新增商品库存管理
     */
    @PreAuthorize("@ss.hasPermi('product:stock:add')")
    @Log(title = "商品库存管理", businessType = BusinessType.INSERT)
    @PostMapping
    public AjaxResult add(@RequestBody Product product)
    {
        return toAjax(productService.insertProduct(product));
    }

    /**
     * 修改商品库存管理
     */
    @PreAuthorize("@ss.hasPermi('product:stock:edit')")
    @Log(title = "商品库存管理", businessType = BusinessType.UPDATE)
    @PutMapping
    public AjaxResult edit(@RequestBody Product product)
    {
        return toAjax(productService.updateProduct(product));
    }

    /**
     * 删除商品库存管理
     */
    @PreAuthorize("@ss.hasPermi('product:stock:remove')")
    @Log(title = "商品库存管理", businessType = BusinessType.DELETE)
	@DeleteMapping("/{ids}")
    public AjaxResult remove(@PathVariable Long[] ids)
    {
        return toAjax(productService.deleteProductByIds(ids));
    }
}
