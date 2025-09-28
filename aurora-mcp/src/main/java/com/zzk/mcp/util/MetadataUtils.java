package com.zzk.mcp.util;

import java.lang.reflect.Field;
import java.util.*;
import java.util.function.Predicate;
import java.util.stream.Collectors;

/**
 * 元数据处理工具类 - 终极版
 * 功能：
 * 1. 自动过滤null/空值字段
 * 2. 支持字段名映射
 * 3. 支持自定义空值判断
 * 4. 支持字段排除
 * 5. 支持字段值转换
 */
public final class MetadataUtils {

    private MetadataUtils() {
        // 工具类禁止实例化
    }

    /**
     * 默认空值判断逻辑
     */
    private static final Predicate<Object> DEFAULT_EMPTY_PREDICATE = value -> {
        if (value == null) {
            return true;
        }
        if (value instanceof CharSequence) {
            return ((CharSequence) value).length() == 0;
        }
        if (value instanceof Collection) {
            return ((Collection<?>) value).isEmpty();
        }
        if (value instanceof Map) {
            return ((Map<?, ?>) value).isEmpty();
        }
        if (value instanceof Object[]) {
            return ((Object[]) value).length == 0;
        }
        return false;
    };
    /**
     * 创建一个谓词，用于判断对象的值是否超过指定长度。
     * 支持类型: CharSequence (如String), Collection, Map, 数组.
     * @param maxLength 最大长度
     * @return 如果值超过最大长度，则返回 true 的谓词
     */
    private static Predicate<Object> lengthExceeds(int maxLength) {
        return value -> {
            if (value == null) {
                return false; // null 值没有长度，不认为它超长
            }
            if (value instanceof CharSequence) {
                return ((CharSequence) value).length() > maxLength;
            }
            if (value instanceof Collection) {
                return ((Collection<?>) value).size() > maxLength;
            }
            if (value instanceof Map) {
                return ((Map<?, ?>) value).size() > maxLength;
            }
            if (value.getClass().isArray()) {
                return java.lang.reflect.Array.getLength(value) > maxLength;
            }
            return false; // 其他类型不检查长度
        };
    }
    /**
     * 将实体对象转换为元数据Map（基础版）
     */
    public static Map<String, Object> convertToMetadata(Object entity) {
        return convertToMetadata(entity, Collections.emptyMap());
    }

    /**
     * 带字段名映射的转换
     * @param fieldNameMapping 字段名映射规则（key=实体字段名，value=目标元数据字段名）
     */
    public static Map<String, Object> convertToMetadata(Object entity, 
                                                       Map<String, String> fieldNameMapping) {
        return convertToMetadata(entity, fieldNameMapping, DEFAULT_EMPTY_PREDICATE);
    }

    /**
     * 完全定制化转换
     * @param fieldNameMapping 字段名映射规则
     * @param isEmptyPredicate 自定义空值判断逻辑
     */
    public static Map<String, Object> convertToMetadata(Object entity,
                                                        Map<String, String> fieldNameMapping,
                                                        Predicate<Object> isEmptyPredicate) {
        if (entity == null) {
            return Collections.emptyMap();
        }

        // 组合谓词：既不能为空，也不能超过长度限制
        Predicate<Object> finalFilter = isEmptyPredicate.negate() // 1. 非空
                .and(lengthExceeds(50).negate()); // 2. 并且长度不超过50

        return Arrays.stream(entity.getClass().getDeclaredFields())
                .filter(field -> {
                    try {
                        field.setAccessible(true);
                        Object value = field.get(entity);
                        // 使用我们组合好的最终过滤器
                        return finalFilter.test(value);
                    } catch (IllegalAccessException e) {
                        return false;
                    }
                })
                .collect(Collectors.toMap(
                        field -> fieldNameMapping.getOrDefault(field.getName(), field.getName()),
                        field -> {
                            try {
                                return convertFieldValue(field.get(entity));
                            } catch (IllegalAccessException e) {
                                return null;
                            }
                        },
                        (oldVal, newVal) -> newVal,
                        LinkedHashMap::new
                ));
    }

    /**
     * 带字段排除的转换
     */
    public static Map<String, Object> convertToMetadataExclude(Object entity, 
                                                             String... excludeFields) {
        Map<String, Object> metadata = convertToMetadata(entity);
        Arrays.stream(excludeFields).forEach(metadata::remove);
        return metadata;
    }

    /**
     * 字段值转换（可被子类重写）
     */
    protected static Object convertFieldValue(Object value) {
        if (value instanceof Collection) {
            return new ArrayList<>((Collection<?>) value);
        }
        if (value instanceof Map) {
            return new LinkedHashMap<>((Map<?, ?>) value);
        }
        if (value instanceof Object[]) {
            return Arrays.asList((Object[]) value);
        }
        return value;
    }

    // ---------- 扩展工具方法 ----------

    /**
     * 转换为下划线风格的元数据字段名
     */
    public static Map<String, Object> convertToMetadataWithSnakeCase(Object entity) {
        return convertToMetadata(entity, field -> {
            String regex = "([a-z])([A-Z]+)";
            String replacement = "$1_$2";
            return field.getName()
                .replaceAll(regex, replacement)
                .toLowerCase();
        });
    }

    /**
     * 带字段名转换函数的版本
     */
    public static Map<String, Object> convertToMetadata(Object entity, 
                                                      java.util.function.Function<Field, String> nameMapper) {
        if (entity == null) {
            return Collections.emptyMap();
        }

        return Arrays.stream(entity.getClass().getDeclaredFields())
                .filter(field -> {
                    try {
                        field.setAccessible(true);
                        return !DEFAULT_EMPTY_PREDICATE.test(field.get(entity));
                    } catch (IllegalAccessException e) {
                        return false;
                    }
                })
                .collect(Collectors.toMap(
                        nameMapper,
                        field -> {
                            try {
                                return convertFieldValue(field.get(entity));
                            } catch (IllegalAccessException e) {
                                return null;
                            }
                        },
                        (oldVal, newVal) -> newVal,
                        LinkedHashMap::new
                ));
    }
}