-- EchoTalk 会员 SKU 播种（内测联调用）
-- 用法：mysql -h127.0.0.1 -uechotalk -p<密码> echotalk < docs/samples/product.seed.sql
-- 价格单位：分。type: 1订阅 2内容包。status: 0下架 1上架。
-- Day7 管理端接管 SKU 增删改后，本文件仅作为初始播种保留。

INSERT INTO products
    (name, description, price, original_price, duration_days, type, status, sort, created_at, updated_at)
VALUES
    ('月度会员', '解锁全部付费内容，有效期30天', 1900, 2900, 30,  1, 1, 10, NOW(), NOW()),
    ('年度会员', '解锁全部付费内容，有效期365天', 12800, 34800, 365, 1, 1, 20, NOW(), NOW());
