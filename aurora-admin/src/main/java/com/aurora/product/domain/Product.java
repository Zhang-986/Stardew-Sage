package com.aurora.product.domain;

import org.apache.commons.lang3.builder.ToStringBuilder;
import org.apache.commons.lang3.builder.ToStringStyle;
import com.aurora.common.annotation.Excel;
import com.aurora.common.core.domain.BaseEntity;

/**
 * 商品库存管理对象 product
 * 
 * @author aurora
 * @date 2025-08-05
 */
public class Product extends BaseEntity
{
    private static final long serialVersionUID = 1L;


    /** 商品ID */
    private Long id;

    /** 商品名称 */
    @Excel(name = "商品名称")
    private String name;

    /** 分类ID */
    @Excel(name = "分类ID")
    private Long categoryId;

    /** 商品价格 */
    @Excel(name = "商品价格")
    private Double price;

    /** 库存数量 */
    @Excel(name = "库存数量")
    private Integer stock;

    /** 商品状态#product_status */
    @Excel(name = "商品状态#product_status")
    private Integer status;

    /** 商品图片URL */
    @Excel(name = "商品图片URL")
    private String imageUrl;

    /** 商品描述 */
    @Excel(name = "商品描述")
    private String description;

    public void setId(Long id) 
    {
        this.id = id;
    }

    public Long getId() 
    {
        return id;
    }

    public void setName(String name) 
    {
        this.name = name;
    }

    public String getName() 
    {
        return name;
    }

    public void setCategoryId(Long categoryId) 
    {
        this.categoryId = categoryId;
    }

    public Long getCategoryId() 
    {
        return categoryId;
    }

    public void setPrice(Double price) 
    {
        this.price = price;
    }

    public Double getPrice() 
    {
        return price;
    }

    public void setStock(Integer stock) 
    {
        this.stock = stock;
    }

    public Integer getStock() 
    {
        return stock;
    }

    public void setStatus(Integer status) 
    {
        this.status = status;
    }

    public Integer getStatus() 
    {
        return status;
    }

    public void setImageUrl(String imageUrl) 
    {
        this.imageUrl = imageUrl;
    }

    public String getImageUrl() 
    {
        return imageUrl;
    }

    public void setDescription(String description) 
    {
        this.description = description;
    }

    public String getDescription() 
    {
        return description;
    }

    @Override
    public String toString() {
        return new ToStringBuilder(this,ToStringStyle.MULTI_LINE_STYLE)
            .append("id", getId())
            .append("name", getName())
            .append("categoryId", getCategoryId())
            .append("price", getPrice())
            .append("stock", getStock())
            .append("status", getStatus())
            .append("imageUrl", getImageUrl())
            .append("description", getDescription())
            .toString();
    }
}
