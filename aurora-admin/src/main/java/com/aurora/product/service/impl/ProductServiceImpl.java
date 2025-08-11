package com.aurora.product.service.impl;

import java.util.List;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import com.aurora.product.mapper.ProductMapper;
import com.aurora.product.domain.Product;
import com.aurora.product.service.IProductService;

/**
 * 商品库存管理Service业务层处理
 * 
 * @author aurora
 * @date 2025-08-05
 */
@Service
public class ProductServiceImpl implements IProductService 
{
    @Autowired
    private ProductMapper productMapper;

    /**
     * 查询商品库存管理
     * 
     * @param id 商品库存管理主键
     * @return 商品库存管理
     */
    @Override
    public Product selectProductById(Long id)
    {
        return productMapper.selectProductById(id);
    }

    /**
     * 查询商品库存管理列表
     * 
     * @param product 商品库存管理
     * @return 商品库存管理
     */
    @Override
    public List<Product> selectProductList(Product product)
    {
        return productMapper.selectProductList(product);
    }

    /**
     * 新增商品库存管理
     * 
     * @param product 商品库存管理
     * @return 结果
     */
    @Override
    public int insertProduct(Product product)
    {
        return productMapper.insertProduct(product);
    }

    /**
     * 修改商品库存管理
     * 
     * @param product 商品库存管理
     * @return 结果
     */
    @Override
    public int updateProduct(Product product)
    {
        return productMapper.updateProduct(product);
    }

    /**
     * 批量删除商品库存管理
     * 
     * @param ids 需要删除的商品库存管理主键
     * @return 结果
     */
    @Override
    public int deleteProductByIds(Long[] ids)
    {
        return productMapper.deleteProductByIds(ids);
    }

    /**
     * 删除商品库存管理信息
     * 
     * @param id 商品库存管理主键
     * @return 结果
     */
    @Override
    public int deleteProductById(Long id)
    {
        return productMapper.deleteProductById(id);
    }
}
