package com.aurora.product.mapper;

import java.util.List;
import com.aurora.product.domain.Product;

/**
 * 商品库存管理Mapper接口
 * 
 * @author aurora
 * @date 2025-08-04
 */
public interface ProductMapper 
{
    /**
     * 查询商品库存管理
     * 
     * @param id 商品库存管理主键
     * @return 商品库存管理
     */
    public Product selectProductById(Long id);

    /**
     * 查询商品库存管理列表
     * 
     * @param product 商品库存管理
     * @return 商品库存管理集合
     */
    public List<Product> selectProductList(Product product);

    /**
     * 新增商品库存管理
     * 
     * @param product 商品库存管理
     * @return 结果
     */
    public int insertProduct(Product product);

    /**
     * 修改商品库存管理
     * 
     * @param product 商品库存管理
     * @return 结果
     */
    public int updateProduct(Product product);

    /**
     * 删除商品库存管理
     * 
     * @param id 商品库存管理主键
     * @return 结果
     */
    public int deleteProductById(Long id);

    /**
     * 批量删除商品库存管理
     * 
     * @param ids 需要删除的数据主键集合
     * @return 结果
     */
    public int deleteProductByIds(Long[] ids);
}
